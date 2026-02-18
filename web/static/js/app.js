function toast(msg, type) {
    var colors = {
        success: "linear-gradient(135deg, #10b981, #059669)",
        error: "linear-gradient(135deg, #ef4444, #dc2626)",
        warning: "linear-gradient(135deg, #f59e0b, #d97706)",
        info: "linear-gradient(135deg, #7c3aed, #6d28d9)"
    };
    Toastify({
        text: msg,
        duration: 2500,
        gravity: "top",
        position: "right",
        style: {
            background: colors[type] || colors.info,
            borderRadius: "6px",
            fontWeight: "500",
            fontSize: "13px",
            padding: "10px 18px",
            boxShadow: "0 4px 20px rgba(0,0,0,.4)"
        }
    }).showToast();
}

function arrayToBase64(arr) {
    var bytes = new Uint8Array(arr);
    var binary = "";
    for (var i = 0; i < bytes.length; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
}

function base64ToArray(b64) {
    var binary = atob(b64);
    var bytes = new Uint8Array(binary.length);
    for (var i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
}

function CloudSaveManager(romName, apiBase) {
    this.romName = romName;
    this.apiBase = apiBase;
    this.emulator = null;
    this.autoSaveTimer = null;
    this.srmPath = null;
}

CloudSaveManager.prototype.init = function () {
    var self = this;
    toast("Aguardando emulador...", "info");
    this.waitForEmulator().then(function (emu) {
        self.emulator = emu;
        toast("Emulador pronto", "success");
        self.srmPath = self.findSRM();
        self.checkCloudSave();
        self.startAutoSave();
    });
};

CloudSaveManager.prototype.waitForEmulator = function () {
    return new Promise(function (resolve) {
        var check = setInterval(function () {
            var emu = window.EJS_emulator;
            if (emu && emu.gameManager) {
                clearInterval(check);
                resolve(emu);
            }
        }, 1000);
    });
};

CloudSaveManager.prototype.findSRM = function () {
    var dirs = ["/data/saves", "/home/web_user/retroarch/userdata/saves"];
    for (var d = 0; d < dirs.length; d++) {
        try {
            var files = this.emulator.Module.FS.readdir(dirs[d]);
            for (var f = 0; f < files.length; f++) {
                if (files[f].endsWith(".srm")) {
                    return dirs[d] + "/" + files[f];
                }
            }
        } catch (e) {}
    }
    return null;
};

CloudSaveManager.prototype.getSRAMData = function () {
    if (this.srmPath) {
        try {
            return this.emulator.Module.FS.readFile(this.srmPath);
        } catch (e) {}
    }
    try {
        var data = this.emulator.gameManager.getSave();
        if (data && data.length > 0) return data;
    } catch (e) {}
    return null;
};

CloudSaveManager.prototype.writeSRAM = function (data) {
    var baseName = this.romName.replace(/\.[^.]+$/, "");
    var dirs = ["/data/saves", "/home/web_user/retroarch/userdata/saves"];
    for (var d = 0; d < dirs.length; d++) {
        try {
            this.emulator.Module.FS.readdir(dirs[d]);
            this.emulator.Module.FS.writeFile(dirs[d] + "/" + baseName + ".srm", data);
            this.srmPath = dirs[d] + "/" + baseName + ".srm";
            return true;
        } catch (e) {}
    }
    try {
        this.emulator.gameManager.loadSave(data);
        return true;
    } catch (e) {}
    return false;
};

CloudSaveManager.prototype.checkCloudSave = function () {
    var self = this;
    var url = this.apiBase + "/saves/" + encodeURIComponent(this.romName) + "/data?type=sram&slot=0";
    fetch(url).then(function (resp) {
        if (!resp.ok) return;
        Swal.fire({
            title: "Save encontrado na nuvem",
            text: "Deseja carregar o save salvo anteriormente?",
            icon: "question",
            showCancelButton: true,
            confirmButtonText: "Carregar",
            cancelButtonText: "Ignorar",
            background: "#18181b",
            color: "#e4e4e7",
            confirmButtonColor: "#7c3aed",
            cancelButtonColor: "#27272a"
        }).then(function (result) {
            if (result.isConfirmed) {
                self.downloadSRAM();
            }
        });
    }).catch(function () {});
};

CloudSaveManager.prototype.upload = function (type, slot, data, silent) {
    var b64 = arrayToBase64(data);
    var url = this.apiBase + "/saves/" + encodeURIComponent(this.romName);
    return fetch(url, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({save_type: type, slot: slot, data: b64})
    }).then(function (resp) {
        if (resp.ok && !silent) {
            var label = type === "sram" ? "SRAM" : "State " + slot;
            toast(label + " salvo na nuvem", "success");
        } else if (!resp.ok && !silent) {
            toast("Erro ao enviar para nuvem", "error");
        }
        return resp.ok;
    }).catch(function () {
        if (!silent) toast("Erro de conexao", "error");
        return false;
    });
};

CloudSaveManager.prototype.download = function (type, slot) {
    var url = this.apiBase + "/saves/" + encodeURIComponent(this.romName) +
        "/data?type=" + type + "&slot=" + slot;
    return fetch(url).then(function (resp) {
        if (!resp.ok) {
            toast("Nenhum save na nuvem", "warning");
            return null;
        }
        return resp.json().then(function (json) {
            return base64ToArray(json.data);
        });
    }).catch(function () {
        toast("Erro ao baixar da nuvem", "error");
        return null;
    });
};

CloudSaveManager.prototype.uploadSRAM = function () {
    var data = this.getSRAMData();
    if (!data || data.length === 0) {
        toast("Nenhum SRAM no emulador", "warning");
        return;
    }
    this.upload("sram", 0, data, false);
};

CloudSaveManager.prototype.downloadSRAM = function () {
    var self = this;
    this.download("sram", 0).then(function (data) {
        if (!data) return;
        if (self.writeSRAM(data)) {
            toast("SRAM carregado da nuvem", "success");
        } else {
            toast("Erro ao injetar SRAM", "error");
        }
    });
};

CloudSaveManager.prototype.saveState = function (slot) {
    try {
        var state = this.emulator.gameManager.getState();
        if (!state || state.length === 0) {
            toast("Falha ao capturar state", "warning");
            return;
        }
        this.upload("state", slot, state, false);
    } catch (e) {
        toast("Erro: " + e.message, "error");
    }
};

CloudSaveManager.prototype.loadState = function (slot) {
    var self = this;
    this.download("state", slot).then(function (data) {
        if (!data) return;
        try {
            self.emulator.gameManager.loadState(data);
            toast("State " + slot + " carregado", "success");
        } catch (e) {
            toast("Erro: " + e.message, "error");
        }
    });
};

CloudSaveManager.prototype.startAutoSave = function () {
    var self = this;
    this.autoSaveTimer = setInterval(function () {
        var data = self.getSRAMData();
        if (data && data.length > 0) {
            self.upload("sram", 0, data, true).then(function (ok) {
                var el = document.getElementById("autoSaveStatus");
                if (!el) return;
                if (ok) {
                    el.innerHTML =
                        '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>' +
                        " Salvo " + new Date().toLocaleTimeString();
                    el.style.color = "#10b981";
                } else {
                    el.innerHTML =
                        '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" x2="12" y1="8" y2="12"/><line x1="12" x2="12.01" y1="16" y2="16"/></svg>' +
                        " Erro";
                    el.style.color = "#ef4444";
                }
            });
        }
    }, 30000);
};

CloudSaveManager.prototype.destroy = function () {
    if (this.autoSaveTimer) clearInterval(this.autoSaveTimer);
};
