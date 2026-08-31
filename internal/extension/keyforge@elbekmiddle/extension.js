// KeyForge GNOME Shell extension.
//
// Runs inside gnome-shell, where global.display.focus_window is available,
// and republishes it as a small D-Bus service so the KeyForge daemon
// (running as a normal, unprivileged process) can ask "what app is
// focused right now?" without relying on org.gnome.Shell.Eval or
// org.gnome.Shell.Introspect (both restricted/unavailable in recent
// GNOME/Wayland sessions).
//
// This extension does no remapping itself — it is a read-only window
// into GNOME Shell state.

import Gio from 'gi://Gio';
import {Extension} from 'resource:///org/gnome/shell/extensions/extension.js';

const BUS_NAME = 'com.elbekmiddle.KeyForge';
const OBJECT_PATH = '/com/elbekmiddle/KeyForge';

const IFACE_XML = `
<node>
  <interface name="com.elbekmiddle.KeyForge">
    <method name="GetActiveApplication">
      <arg type="s" direction="out" name="json" />
    </method>
    <method name="Ping">
      <arg type="s" direction="out" name="pong" />
    </method>
  </interface>
</node>`;

export default class KeyForgeExtension extends Extension {
    enable() {
        this._dbusImpl = Gio.DBusExportedObject.wrapJSObject(IFACE_XML, this);
        this._dbusImpl.export(Gio.DBus.session, OBJECT_PATH);

        this._nameOwnerId = Gio.bus_own_name(
            Gio.BusType.SESSION,
            BUS_NAME,
            Gio.BusNameOwnerFlags.NONE,
            null,
            null,
            null,
        );
    }

    disable() {
        if (this._nameOwnerId) {
            Gio.bus_unown_name(this._nameOwnerId);
            this._nameOwnerId = null;
        }

        if (this._dbusImpl) {
            this._dbusImpl.unexport();
            this._dbusImpl = null;
        }
    }

    // D-Bus method: GetActiveApplication() -> JSON string.
    //
    // Mirrors internal/activeapp.Application on the Go side, minus
    // "executable" — KeyForge resolves that itself from /proc/<pid>/exe,
    // since the shell process usually can't (and shouldn't) read that.
    GetActiveApplication() {
        const window = global.display.focus_window;

        if (!window) {
            return JSON.stringify({});
        }

        let appId = '';

        try {
            appId = window.get_gtk_application_id() || '';
        } catch (e) {
            appId = '';
        }

        return JSON.stringify({
            name: window.get_title() || '',
            class: window.get_wm_class() || '',
            app_id: appId,
            pid: window.get_pid() || 0,
        });
    }

    // D-Bus method: Ping() -> "pong".
    //
    // Lets the daemon cheaply confirm the extension is installed, enabled,
    // and answering on the bus before it starts polling for real data.
    Ping() {
        return 'pong';
    }
}
