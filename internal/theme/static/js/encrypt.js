const _do_decrypt = (encrypted, password) => {
    let key = CryptoJS.enc.Utf8.parse(password);
    let iv = CryptoJS.enc.Utf8.parse(password.substr(0, 16));

    return CryptoJS.AES.decrypt(encrypted, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
    }).toString(CryptoJS.enc.Utf8);
}

const decrypt = (element) => {
    let parent = element.parentNode.parentNode;
    let password = parent.querySelector(".encrypt-password").value;
    let encrypted = parent.querySelector(".encrypt-content").innerText;

    password = CryptoJS.MD5(password).toString()

    let decrypted = "";
    try {
        decrypted = _do_decrypt(encrypted, password);
    } catch (err) {
        console.error(err);
        alert("Failed to decrypt.");
        return;
    }
    if (decrypted == "") {
        alert("Failed to decrypt.");
    } else {
        parent.innerHTML = decrypted;

        let index = -1;
        let elements = document.querySelectorAll(".encrypt-container");
        for (index = 0; index < elements.length; ++index) {
            if (elements[index].isSameNode(parent)) {
                break;
            }
        }
        let storage = sessionStorage;
        let key = location.pathname + ".password." + index;
        storage.setItem(key, password);
    }
}

window.onload = () => {
    let elements = document.querySelectorAll(".encrypt-container");
    elements.forEach((element, index) => {
        let key = location.pathname + ".password." + index;
        let password = sessionStorage.getItem(key);

        if (password) {
            let encrypted = element.querySelector(".encrypt-content").innerText;
            let decrypted = _do_decrypt(encrypted, password);
            element.innerHTML = decrypted;
        } else {
            element.querySelector(".encrypt-form .encrypt-button").addEventListener("click", () => {
                decrypt(element);
            })
            element.querySelector(".encrypt-form .encrypt-password").addEventListener("keypress", (e) => {
                e.keyCode == 13 && decrypt(element);
            })
        }
    });
};