package bonk

const stealthPatchScript = `
(function() {
    var patchFns = [];

    function protectToString(fn, name) {
        patchFns.push({fn: fn, name: name});
    }

    var origToString = Function.prototype.toString;
    var toStringProxy = new Proxy(origToString, {
        apply: function(target, thisArg, args) {
            for (var i = 0; i < patchFns.length; i++) {
                if (thisArg === patchFns[i].fn) {
                    return 'function ' + patchFns[i].name + '() { [native code] }';
                }
            }
            return Reflect.apply(target, thisArg, args);
        }
    });
    Function.prototype.toString = toStringProxy;

    var proto = Object.getPrototypeOf(navigator);
    if (Object.getOwnPropertyDescriptor(proto, 'webdriver')) {
        delete proto.webdriver;
    }
    if (Object.getOwnPropertyDescriptor(navigator, 'webdriver')) {
        delete navigator.webdriver;
    }

    if (navigator.plugins.length === 0) {
        Object.defineProperty(navigator, 'plugins', {
            get: function() {
                var arr = [
                    {name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format', length: 1},
                    {name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: '', length: 1},
                    {name: 'Native Client', filename: 'internal-nacl-plugin', description: '', length: 1}
                ];
                arr.item = function(i) { return this[i] || null; };
                arr.namedItem = function(n) { for (var i=0;i<this.length;i++) if(this[i].name===n) return this[i]; return null; };
                arr.refresh = function() {};
                return arr;
            },
            configurable: true
        });
    }

    Object.defineProperty(navigator, 'languages', {
        get: function() { return ['en-US', 'en']; },
        configurable: true
    });

    if (!window.chrome) window.chrome = {};
    if (!window.chrome.runtime) {
        var rt = {
            connect: function() { return {onMessage: {addListener: function(){}}, postMessage: function(){}, disconnect: function(){}}; },
            sendMessage: function() {}
        };
        protectToString(rt.connect, 'connect');
        protectToString(rt.sendMessage, 'sendMessage');
        window.chrome.runtime = rt;
    }

    if (navigator.permissions && navigator.permissions.query) {
        var origQuery = navigator.permissions.query.bind(navigator.permissions);
        var patchedQuery = function(desc) {
            if (desc && desc.name === 'notifications') {
                return Promise.resolve({state: Notification.permission, onchange: null});
            }
            return origQuery(desc);
        };
        protectToString(patchedQuery, 'query');
        navigator.permissions.query = patchedQuery;
    }

    if (typeof WebGLRenderingContext !== 'undefined') {
        var origGetParam = WebGLRenderingContext.prototype.getParameter;
        var patchedGetParam = function(param) {
            if (param === 37445) return 'Intel Inc.';
            if (param === 37446) return 'Intel Iris OpenGL Engine';
            return origGetParam.call(this, param);
        };
        protectToString(patchedGetParam, 'getParameter');
        WebGLRenderingContext.prototype.getParameter = patchedGetParam;
    }

    if (typeof WebGL2RenderingContext !== 'undefined') {
        var origGetParam2 = WebGL2RenderingContext.prototype.getParameter;
        var patchedGetParam2 = function(param) {
            if (param === 37445) return 'Intel Inc.';
            if (param === 37446) return 'Intel Iris OpenGL Engine';
            return origGetParam2.call(this, param);
        };
        protectToString(patchedGetParam2, 'getParameter');
        WebGL2RenderingContext.prototype.getParameter = patchedGetParam2;
    }

    Object.defineProperty(navigator, 'hardwareConcurrency', {
        get: function() { return 8; },
        configurable: true
    });

    Object.defineProperty(navigator, 'deviceMemory', {
        get: function() { return 8; },
        configurable: true
    });
})();
`
