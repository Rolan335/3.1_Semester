const http = require('http')
const soap = require('soap')
const wsdl = 'https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx?WSDL'


// Запросы к цб рф
async function getValutes() {
    let json = []
    let client = await soap.createClientAsync(wsdl)
    let enumValutes = await client.EnumValutesXMLAsync({ Seld: false })
    enumValutes[0].EnumValutesXMLResult.ValuteData.EnumValutes.forEach((e) => json.push({ code: e.Vcode, name: e.Vname, value: null }));

    let getCurs = await client.GetCursOnDateXMLAsync({ On_date: new Date().toISOString() })
    getCurs[0].GetCursOnDateXMLResult.ValuteData.ValuteCursOnDate.forEach((e) => {
        let i = json.findIndex(obj => obj.name == e.Vname)
        json[i].value = e.Vcurs
    });

    return json
}

async function getValute(args) {
    let json = []
    let client = await soap.createClientAsync(wsdl)
    let getCursDynamicXML = await client.GetCursDynamicXMLAsync(args)
    getCursDynamicXML[0].GetCursDynamicXMLResult.ValuteData.ValuteCursDynamic.forEach((e) => {
        json.push({ date: e.CursDate, value: e.Vcurs })
    })
    return json
}

// Логика нашего сервера
const clientWSDL = require('fs').readFileSync(__dirname + '/service.wsdl', 'utf8');

const service = {
    Service: {
        Port: {
            GetValutes: async function () {
                return { result: await getValutes() }
            },
            GetValute: async function (args) {
                return { result: await getValute(args) }
            }
        }
    }
};

const server = http.createServer((request, response) =>
    response.end('404: Not Found'));

server.listen(8000)

soap.listen(server, '/cbrservice', service, clientWSDL, () =>
    console.log("сервер запущен"));



// // getValutes().then(r => console.log(r))
//// getValute('2023-11-01', '2023-11-20', 'R01565').then(r => console.log(r))


// // async function getValutes() {
// //     let promise = new Promise((resolve, reject) => {
// //         soap.createClientAsync(wsdl, {}, (err, client) => {
// //             let json = [];
// //             client.EnumValutesXML({ Seld: false }, (err, result) => {
// //                 result.EnumValutesXMLResult.ValuteData.EnumValutes.forEach(e => {
// //                     json.push({ code: e.Vcode, name: e.Vname });
// //                     // console.log(json)
// //                 });
// //             });
// //             var x = await client.
// //             console.log(json)
// //         });
// //         resolve(json);
// //     });
// //     let result = await promise
// //     return result
// // }

//// let json = []

//// getValutes()