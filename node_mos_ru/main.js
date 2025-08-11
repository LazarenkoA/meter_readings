const funcs = require('./client');
const {close} = require("./client");
const [,,func,...args] = process.argv;

(async () => {
    if (funcs[func]) {
        await funcs[func](...args);
        await close()

        return
    }

    console.error(`function ${func} not found`);
    process.exit(1);
})();
