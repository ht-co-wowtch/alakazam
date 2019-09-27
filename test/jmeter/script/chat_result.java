import org.json.JSONObject;
import java.nio.*;
import java.util.*;
import java.lang.*;

// 解析Protocol
byte[] rspData = prev.getResponseData();
byte[] packLenByte = Arrays.copyOfRange(rspData,0,4);
byte[] opByte = Arrays.copyOfRange(rspData,6,10);

int packLen = ByteBuffer.wrap(packLenByte).getInt();
int op = ByteBuffer.wrap(opByte).getInt();

String empty = new String();
String failureMessage = new String();

// 判斷Protocol Operation
if (op != 2) {
    failureMessage+="op 錯誤 got: [" + op + "] expect: [2]\n";
}

byte[] jsonData = Arrays.copyOfRange(rspData,10,packLen);
JSONObject body = new JSONObject(new String(jsonData)); 

// 判斷uid
String gotUid = body.getString("uid");
String uid = body.getString("uid");
if (!gotUid.equals(uid)) {
    failureMessage += "uid 錯誤 got: [" + gotUid + "] expect: [" + uid + "]\n";
}

// 判斷key
String gotKey = body.getString("key");
if (!gotKey.matches("[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}")) {
        failureMessage += "key 錯誤 got: [" + gotKey + "]\n";
}

// 判斷status
boolean gotStatus = body.getBoolean("status");
if (!gotStatus) {
    failureMessage += "status 錯誤 got: [" + gotStatus + "] expect: [true]\n";
}

// 判斷room_id
Integer gotRoomId = new Integer(body.getInt("room_id"));
Integer roomId = new Integer(vars.get("room_id"));
if (!gotRoomId.equals(roomId)) {
    failureMessage += "room_id 錯誤 got: [" + gotRoomId + "] expect: [" + roomId + "]\n";
}

// 判斷permission
JSONObject gotPermission = body.getJSONObject("permission");
boolean gotIsMessage = gotPermission.getBoolean("is_message");
boolean gotIsRedEnvelope = gotPermission.getBoolean("is_red_envelope");
if (!gotIsMessage) {
    failureMessage += "permission.is_message 錯誤 got: [" + gotIsMessage + "] expect: [true]\n";
}
if (!gotIsRedEnvelope) {
    failureMessage += "permission.is_red_envelope 錯誤 got: [" + gotIsRedEnvelope + "] expect: [true]\n";
}

// 判斷permission_message
JSONObject gotPermissionMessage = body.getJSONObject("permission_message");
String gotMessage = gotPermissionMessage.getString("is_message");
String gotRedEnvelopeMessage = gotPermissionMessage.getString("is_red_envelope");
if (!gotMessage.equals(empty)) {
    failureMessage += "permission_message.is_message 錯誤 got: [" + gotMessage + "]\n";
}
if (!gotRedEnvelopeMessage.equals(empty)) {
    failureMessage += "permission_message.is_red_envelope 錯誤 got: [" + gotRedEnvelopeMessage + "]\n";
}

if (!failureMessage.equals(empty)) {
    AssertionResult.setFailureMessage(failureMessage);
    AssertionResult.setFailure(true);
}

