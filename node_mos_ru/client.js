
const puppeteer = require('puppeteer');
const fs = require("fs");
const baseURL = 'https://www.mos.ru/';
const readingsURL = 'https://www.mos.ru/services/pokazaniya-vodi-i-tepla/new/';
const deviceURL = 'https://www.mos.ru/api/utility-meter/v1/device?flat=82&payer_code=2290715619';

let browser;
let page;

async function init() {
    if (!browser) {
        browser = await puppeteer.launch({
            headless: true,
            args: [
                '--no-sandbox',
                '--disable-gpu',
                '--proxy-server="direct://"',
                '--proxy-bypass-list=*',
            ]
        });

        page = await browser.newPage();
        await page.setUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36");
        await page.goto(baseURL, {waitUntil: 'networkidle2'});
    }
}


/**
 * @param {import('puppeteer').Page} page
 */
async function login(log, pas) {
    await init();

    let loginButton = await page.$('button[type="button"][aria-label="Вход/Регистрация"]');
    if (!loginButton) {
        throw new Error('Кнопка "Вход/Регистрация" не найдена на странице');
    }

    await Promise.all([
        loginButton.click(),
        page.waitForNavigation({ waitUntil: 'networkidle2' })
    ]);

    loginButton = await page.$('button[id="bind"]');
    if (!loginButton) {
        throw new Error('Кнопка "Войти" не найдена на странице');
    }

    await page.type('input[id="login"]', log);
    await page.type('input[id="password"]', pas);

    await Promise.all([
        loginButton.click(),
        page.waitForNavigation({ waitUntil: 'networkidle2' })
    ]);
}

async function navigateMetersPage(log, pas) {
    await login(log, pas)
    await page.goto(readingsURL, { waitUntil: 'networkidle2' })

    const payerCodeInput = await page.evaluate((keyApartment) => {
        const labels = Array.from(document.querySelectorAll('label'));
        for (const label of labels) {
            if (label.innerText.includes(keyApartment)) {
                const input = label.querySelector('input[type="radio"][name="payerCode"]');
                if (input) {
                    input.click();
                    return true;
                }
            }
        }
        return false;
    }, keyApartment);

    const nextButton = await page.evaluateHandle(() => {
        const buttons = Array.from(document.querySelectorAll('button[type="button"]'));
        return buttons.find(btn => btn.innerText.includes('Продолжить')) || null;
    });
    if (!nextButton) {
        throw new Error('Кнопка "Продолжить" не найдена на странице');
    }

    if(payerCodeInput) {
        await nextButton.click();
    }

    // запрос уйдет через AJAX, ждем изменение контекста
    await page.waitForFunction(() => {
        const el = document.querySelector('h1[data-cy="page-heading"]');
        return el //&& el.innerText.includes('Передача показаний счетчиков');
    }, { timeout: 30000 });
}

async function sendReadingsWater() {


}

/**
 * Получает список счетчиков воды
 *
 * */
async function getMeters(log, pas) {
    await login(log, pas)

    const data = await page.evaluate(async (deviceURL) => {
        const response = await fetch(deviceURL, {
            method: 'GET',
            credentials: 'same-origin', // чтобы отправить куки из текущей сессии
            headers: {
                'Accept': 'application/json',
            }
        });

        if (!response.ok) {
            throw new Error('HTTP error ' + response.status);
        }
        return await response.json();
    }, deviceURL);

    let meterNumbers = [];
    for(const dev of data?.data?.active ?? []) {
        meterNumbers.push(dev?.number)
    }

    console.log(JSON.stringify(meterNumbers));
    return meterNumbers;
}

async function close() {
    if (browser) {
        await browser.close();

        browser = null;
        page = null;
    }
}

module.exports = {
    sendReadingsWater,
    getMeters,
    close,
};