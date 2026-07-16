const curlText = `curl 'https://studio-api.suno.ai/api/generate/v2/' \
  -H 'accept: */*' \
  -H 'authorization: Bearer myToken' \
  -H 'cookie: __client=cookieValue; other=value' \
  --data-raw '{}'`;

const config = {};
const hRegex = /(?:-H|--header)\s+[\$]?(?:(['"])(.*?)\1|([^\s'"]+))/gi;
let match;
while ((match = hRegex.exec(curlText)) !== null) {
    const headerValue = match[2] || match[3];
    if (!headerValue) continue;
    const colonIndex = headerValue.indexOf(':');
    if (colonIndex > -1) {
        const key = headerValue.substring(0, colonIndex).trim().toLowerCase();
        const value = headerValue.substring(colonIndex + 1).trim();
        if (key === 'cookie') config.cookie = value;
    }
}
console.log(config);
