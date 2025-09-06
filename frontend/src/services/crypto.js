/**
 * Cryptographic utility functions
 */

/**
 * Generates MD5 hash of a text string (compatible with Go's MD5)
 * @param {string} text - The text to hash
 * @returns {string} The MD5 hash as a hex-encoded string
 */
export function md5(text) {
  function rotateLeft(value, amount) {
    return (value << amount) | (value >>> (32 - amount));
  }

  function addUnsigned(x, y) {
    const x4 = (x & 0x40000000);
    const y4 = (y & 0x40000000);
    const x8 = (x & 0x80000000);
    const y8 = (y & 0x80000000);
    const result = (x & 0x3FFFFFFF) + (y & 0x3FFFFFFF);
    if (x4 & y4) {
      return (result ^ 0x80000000 ^ x8 ^ y8);
    }
    if (x4 | y4) {
      if (result & 0x40000000) {
        return (result ^ 0xC0000000 ^ x8 ^ y8);
      } else {
        return (result ^ 0x40000000 ^ x8 ^ y8);
      }
    } else {
      return (result ^ x8 ^ y8);
    }
  }

  function md5cmn(q, a, b, x, s, t) {
    return addUnsigned(rotateLeft(addUnsigned(addUnsigned(a, q), addUnsigned(x, t)), s), b);
  }

  function md5ff(a, b, c, d, x, s, t) {
    return md5cmn((b & c) | ((~b) & d), a, b, x, s, t);
  }

  function md5gg(a, b, c, d, x, s, t) {
    return md5cmn((b & d) | (c & (~d)), a, b, x, s, t);
  }

  function md5hh(a, b, c, d, x, s, t) {
    return md5cmn(b ^ c ^ d, a, b, x, s, t);
  }

  function md5ii(a, b, c, d, x, s, t) {
    return md5cmn(c ^ (b | (~d)), a, b, x, s, t);
  }

  function convertToWordArray(string) {
    let wordArray = [];
    let messageLength = string.length;
    let numberOfWords = ((messageLength + 8) - ((messageLength + 8) % 64)) / 64;
    let totalLength = (numberOfWords + 1) * 16;
    let wordIndex = 0;
    let byteIndex = 0;
    
    while (byteIndex < messageLength) {
      let wordCount = (byteIndex - (byteIndex % 4)) / 4;
      let byteCount = (byteIndex % 4) * 8;
      wordArray[wordCount] = (wordArray[wordCount] || 0) | (string.charCodeAt(byteIndex) << byteCount);
      byteIndex++;
    }
    
    let wordCount = (byteIndex - (byteIndex % 4)) / 4;
    let byteCount = (byteIndex % 4) * 8;
    wordArray[wordCount] = wordArray[wordCount] | (0x80 << byteCount);
    wordArray[totalLength - 2] = messageLength << 3;
    wordArray[totalLength - 1] = messageLength >>> 29;
    
    return wordArray;
  }

  function wordToHex(value) {
    let hex = "";
    let byte;
    for (let count = 0; count <= 3; count++) {
      byte = (value >>> (count * 8)) & 255;
      hex += ("0" + byte.toString(16)).slice(-2);
    }
    return hex;
  }

  let x = convertToWordArray(text);
  let a = 0x67452301;
  let b = 0xEFCDAB89;
  let c = 0x98BADCFE;
  let d = 0x10325476;

  for (let k = 0; k < x.length; k += 16) {
    let AA = a;
    let BB = b;
    let CC = c;
    let DD = d;

    a = md5ff(a, b, c, d, x[k + 0], 7, 0xD76AA478);
    d = md5ff(d, a, b, c, x[k + 1], 12, 0xE8C7B756);
    c = md5ff(c, d, a, b, x[k + 2], 17, 0x242070DB);
    b = md5ff(b, c, d, a, x[k + 3], 22, 0xC1BDCEEE);
    a = md5ff(a, b, c, d, x[k + 4], 7, 0xF57C0FAF);
    d = md5ff(d, a, b, c, x[k + 5], 12, 0x4787C62A);
    c = md5ff(c, d, a, b, x[k + 6], 17, 0xA8304613);
    b = md5ff(b, c, d, a, x[k + 7], 22, 0xFD469501);
    a = md5ff(a, b, c, d, x[k + 8], 7, 0x698098D8);
    d = md5ff(d, a, b, c, x[k + 9], 12, 0x8B44F7AF);
    c = md5ff(c, d, a, b, x[k + 10], 17, 0xFFFF5BB1);
    b = md5ff(b, c, d, a, x[k + 11], 22, 0x895CD7BE);
    a = md5ff(a, b, c, d, x[k + 12], 7, 0x6B901122);
    d = md5ff(d, a, b, c, x[k + 13], 12, 0xFD987193);
    c = md5ff(c, d, a, b, x[k + 14], 17, 0xA679438E);
    b = md5ff(b, c, d, a, x[k + 15], 22, 0x49B40821);

    a = md5gg(a, b, c, d, x[k + 1], 5, 0xF61E2562);
    d = md5gg(d, a, b, c, x[k + 6], 9, 0xC040B340);
    c = md5gg(c, d, a, b, x[k + 11], 14, 0x265E5A51);
    b = md5gg(b, c, d, a, x[k + 0], 20, 0xE9B6C7AA);
    a = md5gg(a, b, c, d, x[k + 5], 5, 0xD62F105D);
    d = md5gg(d, a, b, c, x[k + 10], 9, 0x2441453);
    c = md5gg(c, d, a, b, x[k + 15], 14, 0xD8A1E681);
    b = md5gg(b, c, d, a, x[k + 4], 20, 0xE7D3FBC8);
    a = md5gg(a, b, c, d, x[k + 9], 5, 0x21E1CDE6);
    d = md5gg(d, a, b, c, x[k + 14], 9, 0xC33707D6);
    c = md5gg(c, d, a, b, x[k + 3], 14, 0xF4D50D87);
    b = md5gg(b, c, d, a, x[k + 8], 20, 0x455A14ED);
    a = md5gg(a, b, c, d, x[k + 13], 5, 0xA9E3E905);
    d = md5gg(d, a, b, c, x[k + 2], 9, 0xFCEFA3F8);
    c = md5gg(c, d, a, b, x[k + 7], 14, 0x676F02D9);
    b = md5gg(b, c, d, a, x[k + 12], 20, 0x8D2A4C8A);

    a = md5hh(a, b, c, d, x[k + 5], 4, 0xFFFA3942);
    d = md5hh(d, a, b, c, x[k + 8], 11, 0x8771F681);
    c = md5hh(c, d, a, b, x[k + 11], 16, 0x6D9D6122);
    b = md5hh(b, c, d, a, x[k + 14], 23, 0xFDE5380C);
    a = md5hh(a, b, c, d, x[k + 1], 4, 0xA4BEEA44);
    d = md5hh(d, a, b, c, x[k + 4], 11, 0x4BDECFA9);
    c = md5hh(c, d, a, b, x[k + 7], 16, 0xF6BB4B60);
    b = md5hh(b, c, d, a, x[k + 10], 23, 0xBEBFBC70);
    a = md5hh(a, b, c, d, x[k + 13], 4, 0x289B7EC6);
    d = md5hh(d, a, b, c, x[k + 0], 11, 0xEAA127FA);
    c = md5hh(c, d, a, b, x[k + 3], 16, 0xD4EF3085);
    b = md5hh(b, c, d, a, x[k + 6], 23, 0x4881D05);
    a = md5hh(a, b, c, d, x[k + 9], 4, 0xD9D4D039);
    d = md5hh(d, a, b, c, x[k + 12], 11, 0xE6DB99E5);
    c = md5hh(c, d, a, b, x[k + 15], 16, 0x1FA27CF8);
    b = md5hh(b, c, d, a, x[k + 2], 23, 0xC4AC5665);

    a = md5ii(a, b, c, d, x[k + 0], 6, 0xF4292244);
    d = md5ii(d, a, b, c, x[k + 7], 10, 0x432AFF97);
    c = md5ii(c, d, a, b, x[k + 14], 15, 0xAB9423A7);
    b = md5ii(b, c, d, a, x[k + 5], 21, 0xFC93A039);
    a = md5ii(a, b, c, d, x[k + 12], 6, 0x655B59C3);
    d = md5ii(d, a, b, c, x[k + 3], 10, 0x8F0CCC92);
    c = md5ii(c, d, a, b, x[k + 10], 15, 0xFFEFF47D);
    b = md5ii(b, c, d, a, x[k + 1], 21, 0x85845DD1);
    a = md5ii(a, b, c, d, x[k + 8], 6, 0x6FA87E4F);
    d = md5ii(d, a, b, c, x[k + 15], 10, 0xFE2CE6E0);
    c = md5ii(c, d, a, b, x[k + 6], 15, 0xA3014314);
    b = md5ii(b, c, d, a, x[k + 13], 21, 0x4E0811A1);
    a = md5ii(a, b, c, d, x[k + 4], 6, 0xF7537E82);
    d = md5ii(d, a, b, c, x[k + 11], 10, 0xBD3AF235);
    c = md5ii(c, d, a, b, x[k + 2], 15, 0x2AD7D2BB);
    b = md5ii(b, c, d, a, x[k + 9], 21, 0xEB86D391);

    a = addUnsigned(a, AA);
    b = addUnsigned(b, BB);
    c = addUnsigned(c, CC);
    d = addUnsigned(d, DD);
  }

  return (wordToHex(a) + wordToHex(b) + wordToHex(c) + wordToHex(d)).toLowerCase();
}

export default {
  md5
};
