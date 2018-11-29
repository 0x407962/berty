package chat.berty.ble;

import android.annotation.SuppressLint;
import android.annotation.TargetApi;
import android.bluetooth.BluetoothGattCharacteristic;
import android.bluetooth.BluetoothGattService;
import android.os.Build;
import android.util.Log;

import java.util.UUID;

import static android.bluetooth.BluetoothGattCharacteristic.PERMISSION_READ;
import static android.bluetooth.BluetoothGattCharacteristic.PERMISSION_WRITE;
import static android.bluetooth.BluetoothGattCharacteristic.PROPERTY_READ;
import static android.bluetooth.BluetoothGattCharacteristic.PROPERTY_WRITE;
import static android.bluetooth.BluetoothGattService.SERVICE_TYPE_PRIMARY;

@SuppressLint("LongLogTag")
@TargetApi(Build.VERSION_CODES.JELLY_BEAN_MR2)
public class BertyConstants {
    final static int BLUETOOTH_ENABLE_REQUEST = 1;
    final static UUID SERVICE_UUID = UUID.fromString("A06C6AB8-886F-4D56-82FC-2CF8610D6664");
    final static UUID WRITER_UUID = UUID.fromString("000CBD77-8D30-4EFF-9ADD-AC5F10C2CC1C");
    final static UUID CLOSER_UUID = UUID.fromString("AD127A46-D065-4D72-B15A-EB2B3DA20561");
    final static UUID IS_READY_UUID = UUID.fromString("D27DE0B5-2170-4C59-9C0B-750C760C74E6");
    final static UUID MA_UUID = UUID.fromString("9B827770-DC72-4C55-B8AE-0870C7AC15A8");
    final static UUID PEER_ID_UUID = UUID.fromString("0EF50D30-E208-4315-B323-D05E0A23E6B3");
    final static UUID ACCEPT_UUID = UUID.fromString("6F110ECA-9FCC-4BB3-AB45-6F13565E2E34");
    final static BluetoothGattService mService = new BluetoothGattService(SERVICE_UUID, SERVICE_TYPE_PRIMARY);
    final static BluetoothGattCharacteristic acceptCharacteristic = new BluetoothGattCharacteristic(ACCEPT_UUID, PROPERTY_READ | PROPERTY_WRITE, PERMISSION_READ | PERMISSION_WRITE);
    final static BluetoothGattCharacteristic maCharacteristic = new BluetoothGattCharacteristic(MA_UUID, PROPERTY_READ, PERMISSION_READ);
    final static BluetoothGattCharacteristic peerIDCharacteristic = new BluetoothGattCharacteristic(PEER_ID_UUID, PROPERTY_READ, PERMISSION_READ);
    final static BluetoothGattCharacteristic writerCharacteristic = new BluetoothGattCharacteristic(WRITER_UUID, PROPERTY_WRITE, PERMISSION_WRITE);
    final static BluetoothGattCharacteristic isRdyCharacteristic = new BluetoothGattCharacteristic(IS_READY_UUID, PROPERTY_WRITE, PERMISSION_WRITE);
    final static BluetoothGattCharacteristic closerCharacteristic = new BluetoothGattCharacteristic(CLOSER_UUID, PROPERTY_WRITE, PERMISSION_WRITE);
    private static final String TAG = "chat.berty.ble.BertyConstants";

    public static BluetoothGattService createService() {
        Log.e(TAG, "createService()");

            if (!mService.addCharacteristic(acceptCharacteristic) ||
                !mService.addCharacteristic(maCharacteristic) ||
            !mService.addCharacteristic(peerIDCharacteristic) ||
            !mService.addCharacteristic(writerCharacteristic) ||
            !mService.addCharacteristic(isRdyCharacteristic) ||
            !mService.addCharacteristic(closerCharacteristic)) {
                Log.e(TAG, "Error adding characteristic");
            }

            return mService;
    }
}
