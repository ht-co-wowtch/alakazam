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
if (op != 4) {
    failureMessage+="op 錯誤 got: [" + op + "] expect: [4]\n";
}

if (!failureMessage.equals(empty)) {
    AssertionResult.setFailureMessage(failureMessage);
    AssertionResult.setFailure(true);
}
