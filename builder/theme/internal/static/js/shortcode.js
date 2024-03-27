var shortcodes = {
    loadJS: (src, useAsync = true, useDefer = false) => {
        let element = document.createElement("script");
        element.src = src;
        element.type = "text/javascript";
        element.async = useAsync;
        element.defer = useDefer;

        document.body.appendChild(element);
    },
    loadCSS: (src) => {
        let element = document.createElement("link");
        element.rel = "stylesheet";
        element.type = "text/css";
        element.href = src

        document.getElementsByTagName("head")[0].appendChild(element);
    }
};