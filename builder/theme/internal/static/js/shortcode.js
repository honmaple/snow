var shortcodes = {
    loadJS: (src, useAsync = true) => {
        let element = document.createElement("script");
        element.setAttribute("src", src);
        element.setAttribute("type", "text/javascript");
        element.setAttribute("async", useAsync);

        document.body.appendChild(element);
    },
    loadCSS: (src) => {
        let element = document.createElement("link");
        element.setAttribute("rel", "stylesheet");
        element.setAttribute("type", "text/css");
        element.setAttribute("href", src);

        document.getElementsByTagName("head")[0].appendChild(element);
    }
};