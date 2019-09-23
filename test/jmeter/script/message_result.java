import org.json.JSONObject;
import java.nio.*;
import java.util.*;
import java.lang.*;

// 解析Protocol
byte[] rspData = prev.getResponseData();

byte[] titlePackLenByte = Arrays.copyOfRange(rspData,0,4);
byte[] titleOpByte = Arrays.copyOfRange(rspData,6,10);

int titlePackLen = ByteBuffer.wrap(titlePackLenByte).getInt();
int titleOp = ByteBuffer.wrap(titleOpByte).getInt();

String empty = new String();
String failureMessage = new String();

// 判斷Protocol Operation
switch(titleOp) { 
    case 5: 
        for (int offset = 10; offset < rspData.length; offset += titlePackLen) {
            byte[] packLenByte = Arrays.copyOfRange(rspData, 0+offset, 4+offset);
            byte[] opByte = Arrays.copyOfRange(rspData, 6+offset, 10+offset);

            int packLen = ByteBuffer.wrap(packLenByte).getInt();
            int op = ByteBuffer.wrap(opByte).getInt();

            if (op != 6) {
                failureMessage+="op 錯誤 got: [" + op + "] expect: [6]\n";
                break;
            }

            byte[] jsonData = Arrays.copyOfRange(rspData, offset + 10, offset + packLen);
            JSONObject body = new JSONObject(new String(jsonData)); 

            int id = body.getInt("id");

            if (id == 0) {
                failureMessage+="id 錯誤 " + new String(jsonData) + "\n";
                break;
            }
        }
        break; 
    default: 
    failureMessage+="op 錯誤 got: [" + op + "] expect: [5]\n";
}

if (!failureMessage.equals(empty)) {
    AssertionResult.setFailureMessage(failureMessage);
    AssertionResult.setFailure(true);
}