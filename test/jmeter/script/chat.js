load("encoding-indexes.js")
load("encoding.js")

var token = '{"token":"' + vars.get("token") + '", "room_id":' + vars.get("room_id") + '}';
var headerBuf = new ArrayBuffer(10);
var headerView = new DataView(headerBuf, 0);
var textEncoder = new TextEncoder();
var bodyBuf = textEncoder.encode(token);

headerView.setInt32(0, 10 + bodyBuf.byteLength);
headerView.setInt16(4, 10);
headerView.setInt32(6, 1);

vars.put("chatData", mergeArrayBuffer(headerBuf, bodyBuf))

function mergeArrayBuffer(ab1, ab2) {
   var u81 = new Uint8Array(ab1),
       u82 = new Uint8Array(ab2),
       res = new Uint8Array(ab1.byteLength + ab2.byteLength);
   res.set(u81, 0);
   res.set(u82, ab1.byteLength);
   
   return byteToHexString(res);
}

function byteToHexString(uint8arr) {
  if (!uint8arr) {
    return '';
  }
  
  var hexStr = '';
  for (var i = 0; i < uint8arr.length; i++) {
    var hex = (uint8arr[i] & 0xff).toString(16);
    hex = (hex.length === 1) ? '0' + hex : hex;
    hexStr += hex;
  }
  
  return hexStr.toUpperCase();
}



