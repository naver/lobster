$(function() {
    var urlParams = new URLSearchParams(window.location.search);
    var t = new Date();
    var endTime = Math.ceil(t.getTime() / 1000);
    t.setMinutes(t.getMinutes() - 30)
    var startTime = Math.ceil(t.getTime() / 1000);
    var isParamChanged = false;

    if (!urlParams.has('start')) {
        urlParams.set('start', startTime);
        urlParams.set('end', endTime);
        isParamChanged = true
    }
    
    if (isParamChanged) {
        window.location.search = urlParams;
    }

    if (urlParams.has('page')) {
        document.getElementById('pager').value = urlParams.get('page');
    }

    ids.forEach(function(target){
        if (urlParams.has(target) && urlParams.has('namespaces')) {
            document.querySelector('#'.concat(target)).setValue(decodeURIComponent(urlParams.get(target)).split('|'));
        } else {
            urlParams.delete(target);
        }
    });
    
    if (urlParams.has('include')) {
        document.getElementById("include").setValue(decodeURIComponent(urlParams.get('include')));
    }
    if (urlParams.has('exclude')) {
        document.getElementById("exclude").setValue(decodeURIComponent(urlParams.get('exclude')));
    }
});

function refreshRange() {
    var urlParams = new URLSearchParams(window.location.search);
    var t = new Date();
    var endTime = Math.ceil(t.getTime() / 1000);
    t.setMinutes(t.getMinutes() - 30)
    var startTime = Math.ceil(t.getTime() / 1000);
    
    urlParams.set('start', startTime);
    urlParams.set('end', endTime);

    window.location.search = urlParams;
}

function onChange() {
    const urlParams = new URLSearchParams(window.location.search);
    const inputs = this.value.join('|')
    
    if (this.value.length == 0) {
        if (this.name == 'namespaces') {
            clearAll(urlParams)
        } else {
            urlParams.delete(this.name);
        }
    } else {
        urlParams.set(this.name, inputs);
    }
    
    urlParams.set('page', 1)
    window.location.search = urlParams;
}

function onReset() {
    const urlParams = new URLSearchParams(window.location.search);
    
    if (this.name == 'namespaces') {
        clearAll(urlParams)
    } else {
        urlParams.delete(this.name);
    }
    
    urlParams.set('page', 1)
    window.location.search = urlParams;
}

function clearAll(urlParams) {
    ids.forEach(function(target){
        urlParams.delete(target);
    });
}

function setFilter() {
        const urlParams = new URLSearchParams(window.location.search);
        const include = document.getElementById('include').value;
        const exclude = document.getElementById('exclude').value;
        
        if (include) {
            urlParams.set('include', encodeURIComponent(include));
        } else {
            urlParams.delete('include');
        }

        if (exclude) {
            urlParams.set('exclude', encodeURIComponent(exclude));
        } else {
            urlParams.delete('exclude');
        }

        window.location.search = urlParams;
}

function setRange(start, end) {
        const urlParams = new URLSearchParams(window.location.search);
        const startTime = new Date(start.format()).getTime() / 1000
        const endTime = new Date(end.format()).getTime() / 1000
        
        urlParams.set('start', startTime);
        urlParams.set('end', endTime);
        window.location.search = urlParams;
}

function setPage(value) {
        const urlParams = new URLSearchParams(window.location.search);
        
        urlParams.set('page', value);
        window.location.search = urlParams;
        document.getElementById('pager').value = value;
}

window.addEventListener("load", function (event) {
        const urlParams = new URLSearchParams(window.location.search);
        new DateRangePicker('datetimerange',
            {
                startDate: new Date(parseInt(urlParams.get('start')) * 1000),
                endDate: new Date(parseInt(urlParams.get('end')) * 1000),
                timePicker: true,
                autoApply: true,
                opens: 'left',
                ranges: {
                    '1 Hours': [moment().subtract(1,'hours'), moment()],
                    'Today': [moment().startOf('day'), moment().endOf('day')],
                    'Yesterday': [moment().subtract(1, 'days').startOf('day'), moment().subtract(1, 'days').endOf('day')],
                    'Last 7 Days': [moment().subtract(6, 'days').startOf('day'), moment().endOf('day')],
                    'This Month': [moment().startOf('month').startOf('day'), moment().endOf('month').endOf('day')],
                },
                locale: {
                    format: "YYYY-MM-DD HH:mm:ss",
                }
            },setRange)
});
window.addEventListener('apply.daterangepicker', function (ev) {
        setRange(ev.detail.startDate,ev.detail.endDate);
});
