var headerBuf = new ArrayBuffer(10);
var headerView = new DataView(headerBuf, 0);
headerView.setInt32(0, 10);
headerView.setInt16(4, 10);
headerView.setInt32(6, 3);

vars.put("heartbeat", byteToHexString(new Uint8Array(headerBuf)))

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