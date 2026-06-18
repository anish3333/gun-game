import { encode, decode } from 'https://esm.sh/@msgpack/msgpack@3.1.2';

export async function decodeWsData(data) {
  if (typeof data === 'string') {
    return JSON.parse(data);
  }
  const buf = data instanceof ArrayBuffer
    ? new Uint8Array(data)
    : new Uint8Array(await data.arrayBuffer());
  return decode(buf);
}

export function encodeWsMessage(msg, encoding) {
  if (encoding === 'msgpack') {
    return encode(msg);
  }
  return JSON.stringify(msg);
}

export function isBinaryEncoding(encoding) {
  return encoding === 'msgpack';
}
