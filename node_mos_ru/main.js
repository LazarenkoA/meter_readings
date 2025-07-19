const funcs = require('./client');
const {close} = require("./client");
const [,,func,login, password] = process.argv;

(async () => {
    if (funcs[func]) {
        await funcs[func](login, password);
        await close()

        return
    }

    console.error(`function ${func} not found`);
    process.exit(1);
})();
