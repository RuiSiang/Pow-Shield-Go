 // sha256 hash
 async function sha256(val) {
    const h = await crypto.subtle.digest('SHA-256', new TextEncoder('utf-8').encode(val));
    const view = new DataView(h);
    const hexes = [];
    for (let i = 0; i < view.byteLength; i += 4)
        hexes.push(('00000000' + view.getUint32(i).toString(16)).slice(-8));
    return hexes.join('');
}

// proof of work
async function pow(prefix, difficulty) {
    // initialize the nonce with random number 0 to 1000000
    let nonce = Math.floor(Math.random() * 1000000);
    while (true) {
        const hash = await sha256(prefix + nonce);
        if (hash.startsWith('0'.repeat(difficulty)))
            return nonce;
        nonce++;
    }

    return nonce;
}