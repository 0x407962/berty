package tech.berty.core;

import android.content.Intent;
import android.net.Uri;
import android.provider.Settings;

import com.facebook.react.bridge.ReactApplicationContext;
import com.facebook.react.bridge.ReactContextBaseJavaModule;
import com.facebook.react.bridge.Promise;
import com.facebook.react.bridge.ReactMethod;

import tech.berty.core.notification.NotificationNative;
import core.Core;
import core.NativeBridge;
import core.MobileNotification;

public class BertyCoreModule extends ReactContextBaseJavaModule {
    private Logger logger = new Logger("core.berty.tech");

    private ReactApplicationContext reactContext;
    private MobileNotification notificationDriver = Core.getNotificationDriver();
    private ConnectivityUpdateHandler connectivity;
    private NativeBridge daemon;

    public BertyCoreModule(ReactApplicationContext reactContext) {
        super(reactContext);

        this.reactContext = reactContext;

        String storagePath = reactContext.getFilesDir().getAbsolutePath();
        try {
            Core.setStoragePath(storagePath);
        } catch (Exception error) {
            logger.format(Level.ERROR, this.getName(), error.getMessage());
        }

        daemon = Core.newNativeBridge(this.logger);

        this.notificationDriver.setNative(new NotificationNative());

        connectivity = new ConnectivityUpdateHandler(reactContext);
    }

    public String getName() {
        return "BertyCore";
    }

    @ReactMethod
    public void invoke(final String method, final String message, final Promise promise) {
        new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    String data = daemon.invoke(method, message);
                    promise.resolve(data);
                } catch (Exception err) {
                    logger.format(Level.ERROR, getName(), "Invoke daemon failed: %s", err);
                    promise.reject(err);
                }
            }
        }).start();
    }

    @ReactMethod
    public void setCurrentRoute(String route) {
        Core.setAppRoute(route);
    }

    @ReactMethod
    public void throwException() throws Exception {
        throw new Exception("Manually thrown exception");
    }

    @ReactMethod
    public void crash() {
        this.crash();
    }

    @ReactMethod
    public void openSettings() {
        Intent intent = new Intent();
        intent.setAction(Settings.ACTION_APPLICATION_DETAILS_SETTINGS);
        Uri uri = Uri.fromParts("package", reactContext.getPackageName(), null);
        intent.setData(uri);
        reactContext.startActivity(intent);
    }
}
