/*
 * Copyright (c) 2024-present NAVER Corp
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/golang/glog"
	cp "github.com/otiai10/copy"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

var (
	regexpName = regexp.MustCompile(`^(.+)(-[0-9a-zA-Z]{8,10}|)-[0-9a-zA-Z]*$`)
)

func TargetPathAppearance(path string) int {
	return len(strings.Split(path, "/"))
}

func MeasureElapse(f func() string, msg string) {
	start := time.Now()
	result := f()
	elapsed := time.Since(start)

	glog.V(2).Infof("msg: %s | took: %v | result: %s", msg, elapsed, result)
}

func ConvertStringToTimestamp(str string) (Timestamp, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return Timestamp{}, errors.Wrapf(err, "invalid timestamp: %s", str)
	}
	return Timestamp{time.Unix(value, 0)}, nil
	// return time.Unix(value/1e3, (value%1e3)*1e6), nil
}

func LookupEndpoints(domain string) ([]string, error) {
	endpoints := map[string]struct{}{}
	_, addrs, err := net.LookupSRV("", "", domain)
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ips, err := net.LookupIP(addr.Target)
		if err != nil {
			return nil, err
		}
		endpoints[fmt.Sprintf("%s:%d", ips[0], addr.Port)] = struct{}{}
	}

	results := []string{}
	for endpoint := range endpoints {
		results = append(results, endpoint)
	}

	return results, nil
}

func GetLocalAddress() string {
	interfaces, _ := net.Interfaces()

	for _, inter := range interfaces {
		if inter.Name != "bond0" && inter.Name != "eth0" {
			continue
		}

		addrs, err := inter.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if n, ok := addr.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil {
				return n.IP.String()
			}
		}
	}

	return ""
}

func FindSetName(input string) (string, error) {
	ret := regexpName.FindStringSubmatch(input)
	if len(ret) < 2 {
		return "", fmt.Errorf("can't find deployment name in %s", input)
	}
	return ret[1], nil
}

func MoveFileContents(src, dst string) error {
	fSrc, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := fSrc.Close(); err == nil {
			err = cErr
		}
	}()

	fDst, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := fDst.Close(); err == nil {
			err = cErr
		}
	}()

	if _, err = io.Copy(fDst, fSrc); err != nil {
		return err
	}

	if err = fDst.Sync(); err != nil {
		return err
	}

	if err := os.Truncate(src, 0); err != nil {
		return err
	}

	if _, err := fSrc.Seek(0, 0); err != nil {
		return err
	}

	return nil
}

func BytesToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

func Copy(src string, dst string) error {
	return cp.Copy(src, dst, cp.Options{Sync: true})
}

func RemoveDir(dst string) error {
	return os.RemoveAll(dst)
}

func FlushFile(fDst *os.File) error {
	if err := fDst.Truncate(0); err != nil {
		return err
	}

	if _, err := fDst.Seek(0, 0); err != nil {
		return err
	}

	return fDst.Sync()
}

func DiskInfo(path string) (cap uint64, used uint64, err error) {
	fs := unix.Statfs_t{}
	err = unix.Statfs(path, &fs)
	if err != nil {
		return
	}

	cap = fs.Blocks * uint64(fs.Bsize)
	used = cap - fs.Bfree*uint64(fs.Bsize)
	return
}

func WriteFile(path, fileName string, data []byte) error {
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", path, fileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := f.Close(); err == nil {
			err = cErr
		}
	}()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return f.Sync()
}

func CountSubDirectories(dir string) (int, error) {
	n := 0
	subFiles, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	for _, subfile := range subFiles {
		if !subfile.IsDir() {
			continue
		}
		n = n + 1
	}

	return n, nil
}

func NewCertPoolForRootCA(caCert []byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	return pool, nil
}
