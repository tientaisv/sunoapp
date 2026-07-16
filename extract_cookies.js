const fs = require('fs');
const cookies = JSON.parse(fs.readFileSync('/home/data/taonhac/cookies_suno_com_1784099329232.json', 'utf8'));
let cookieStr = cookies
    .filter(c => c.name === '__client' || c.name === '__client_uat')
    .map(c => `${c.name}=${c.value}`)
    .join('; ');
console.log(cookieStr);
