/// reference path="./core-api-guest.d.ts"

console.log("dashboard invocation")

const moc = requireApi('core')

// we use the doT.js template engine to render server-side

// doT.js
// 2011-2014, Laura Doktorova, https://github.com/olado/doT
// Licensed under the MIT license.

var doT = (function () {
    "use strict";

    var doT = {
        name: "doT",
        version: "1.1.1",
        templateSettings: {
            evaluate: /\{\{([\s\S]+?(\}?)+)\}\}/g,
            interpolate: /\{\{=([\s\S]+?)\}\}/g,
            encode: /\{\{!([\s\S]+?)\}\}/g,
            use: /\{\{#([\s\S]+?)\}\}/g,
            useParams: /(^|[^\w$])def(?:\.|\[[\'\"])([\w$\.]+)(?:[\'\"]\])?\s*\:\s*([\w$\.]+|\"[^\"]+\"|\'[^\']+\'|\{[^\}]+\})/g,
            define: /\{\{##\s*([\w\.$]+)\s*(\:|=)([\s\S]+?)#\}\}/g,
            defineParams: /^\s*([\w$]+):([\s\S]+)/,
            conditional: /\{\{\?(\?)?\s*([\s\S]*?)\s*\}\}/g,
            iterate: /\{\{~\s*(?:\}\}|([\s\S]+?)\s*\:\s*([\w$]+)\s*(?:\:\s*([\w$]+))?\s*\}\})/g,
            varname: "it",
            strip: true,
            append: true,
            selfcontained: false,
            doNotSkipEncoded: false
        },
        template: undefined, //fn, compile template
        compile: undefined, //fn, for express
        log: true
    }, _globals;

    doT.encodeHTMLSource = function (doNotSkipEncoded) {
        var encodeHTMLRules = { "&": "&#38;", "<": "&#60;", ">": "&#62;", '"': "&#34;", "'": "&#39;", "/": "&#47;" },
            matchHTML = doNotSkipEncoded ? /[&<>"'\/]/g : /&(?!#?\w+;)|<|>|"|'|\//g;
        return function (code) {
            return code ? code.toString().replace(matchHTML, function (m) { return encodeHTMLRules[m] || m; }) : "";
        };
    };

    var startend = {
        append: { start: "'+(", end: ")+'", startencode: "'+encodeHTML(" },
        split: { start: "';out+=(", end: ");out+='", startencode: "';out+=encodeHTML(" }
    }, skip = /$^/;

    function resolveDefs(c, block, def) {
        return ((typeof block === "string") ? block : block.toString())
            .replace(c.define || skip, function (m, code, assign, value) {
                if (code.indexOf("def.") === 0) {
                    code = code.substring(4);
                }
                if (!(code in def)) {
                    if (assign === ":") {
                        if (c.defineParams) value.replace(c.defineParams, function (m, param, v) {
                            def[code] = { arg: param, text: v };
                        });
                        if (!(code in def)) def[code] = value;
                    } else {
                        new Function("def", "def['" + code + "']=" + value)(def);
                    }
                }
                return "";
            })
            .replace(c.use || skip, function (m, code) {
                if (c.useParams) code = code.replace(c.useParams, function (m, s, d, param) {
                    if (def[d] && def[d].arg && param) {
                        var rw = (d + ":" + param).replace(/'|\\/g, "_");
                        def.__exp = def.__exp || {};
                        def.__exp[rw] = def[d].text.replace(new RegExp("(^|[^\\w$])" + def[d].arg + "([^\\w$])", "g"), "$1" + param + "$2");
                        return s + "def.__exp['" + rw + "']";
                    }
                });
                var v = new Function("def", "return " + code)(def);
                return v ? resolveDefs(c, v, def) : v;
            });
    }

    function unescape(code) {
        return code.replace(/\\('|\\)/g, "$1").replace(/[\r\t\n]/g, " ");
    }

    doT.template = function (tmpl, c, def) {
        c = c || doT.templateSettings;
        var cse = c.append ? startend.append : startend.split, needhtmlencode, sid = 0, indv,
            str = (c.use || c.define) ? resolveDefs(c, tmpl, def || {}) : tmpl;

        str = ("var out='" + (c.strip ? str.replace(/(^|\r|\n)\t* +| +\t*(\r|\n|$)/g, " ")
            .replace(/\r|\n|\t|\/\*[\s\S]*?\*\//g, "") : str)
            .replace(/'|\\/g, "\\$&")
            .replace(c.interpolate || skip, function (m, code) {
                return cse.start + unescape(code) + cse.end;
            })
            .replace(c.encode || skip, function (m, code) {
                needhtmlencode = true;
                return cse.startencode + unescape(code) + cse.end;
            })
            .replace(c.conditional || skip, function (m, elsecase, code) {
                return elsecase ?
                    (code ? "';}else if(" + unescape(code) + "){out+='" : "';}else{out+='") :
                    (code ? "';if(" + unescape(code) + "){out+='" : "';}out+='");
            })
            .replace(c.iterate || skip, function (m, iterate, vname, iname) {
                if (!iterate) return "';} } out+='";
                sid += 1; indv = iname || "i" + sid; iterate = unescape(iterate);
                return "';var arr" + sid + "=" + iterate + ";if(arr" + sid + "){var " + vname + "," + indv + "=-1,l" + sid + "=arr" + sid + ".length-1;while(" + indv + "<l" + sid + "){"
                    + vname + "=arr" + sid + "[" + indv + "+=1];out+='";
            })
            .replace(c.evaluate || skip, function (m, code) {
                return "';" + unescape(code) + "out+='";
            })
            + "';return out;")
            .replace(/\n/g, "\\n").replace(/\t/g, '\\t').replace(/\r/g, "\\r")
            .replace(/(\s|;|\}|^|\{)out\+='';/g, '$1').replace(/\+''/g, "");
        //.replace(/(\s|;|\}|^|\{)out\+=''\+/g,'$1out+=');

        if (needhtmlencode) {
            if (!c.selfcontained && _globals && !_globals._encodeHTML) _globals._encodeHTML = doT.encodeHTMLSource(c.doNotSkipEncoded);
            str = "var encodeHTML = typeof _encodeHTML !== 'undefined' ? _encodeHTML : ("
                + doT.encodeHTMLSource.toString() + "(" + (c.doNotSkipEncoded || '') + "));"
                + str;
        }
        try {
            return new Function(c.varname, str);
        } catch (e) {
            /* istanbul ignore else */
            if (typeof console !== "undefined") console.log("Could not create a template function: " + str);
            throw e;
        }
    };

    doT.compile = function (tmpl, def) {
        return doT.template(tmpl, null, def);
    };

    return doT;
})();

////////// end doT

function getDashboardHtml() {
    var status = JSON.parse(moc.getStatus());

    var plugs = []
    for (var spec in status.plugs) {
        var fs = spec.indexOf("/")
        var method = spec.substr(0, fs)
        var path = spec.substr(fs + 1)
        var p = JSON.parse(status.plugs[spec])

        plugs.push({
            method: method,
            path: path,
            type: p.type,
            name: p.name,
            start_function: p.start_function
        })
    }
    status.plugs = plugs

    var templateTechID = moc.getBlobTechIdFromName("dashboard-template");
    var templateContent = moc.getBlobBytesAsString(templateTechID);

    var tempFn = doT.template(templateContent);
    var resultText = tempFn(status);

    moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "Content-Type", "text/html");
    moc.writeExchangeBuffer(moc.getOutputBufferId(), resultText);

    return 200
}