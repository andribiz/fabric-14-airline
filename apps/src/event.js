const { FileSystemWallet, Gateway } = require("fabric-network");
const fs = require("fs");
const yaml = require("js-yaml");
const { SIGINT } = require("constants");

const CCP =
  "../../network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml";
const WALLET = "./wallet";
const gateway = new Gateway();

async function main() {
  const conn = yaml.safeLoad(fs.readFileSync(CCP, "utf8"));
  const wallet = new FileSystemWallet(WALLET);

  await gateway.connect(conn, {
    wallet,
    identity: "john@org1.example.com",
    discovery: { enabled: true },
  });

  try {
    console.log("Connecting");
    const network = await gateway.getNetwork("airlinechannel");
    const contract = network.getContract("invoicecc");
    console.log("Registering");
    const list = await contract.addContractListener(
      "InvoiceCreatedListener",
      "InvoiceCreated",
      (err, event, blockNumber, transactionId, status) => {
        if (err) {
          console.log(err.message);
          return;
        }
        console.log(
          `Event:${event.payload.toString()} BlockNumber:${blockNumber} TxID: ${transactionId} ${status}`
        );
      }
    );

    network.addBlockListener("BlockListener", (err, block) => {
      if (err) {
        console.log(err);
        return;
      }
      console.log(block.header.number);
    });

    console.log("Listening");
    // setInterval(() => {}, 100);
  } catch (err) {
    console.log(err.message);
  }
  //   gateway.disconnect();
}

process.on("SIGINT", () => {
  console.log("Disconnecting...");
  gateway.disconnect();
});

main();

// } else if (process.argv[2] === "contractListener") {
// const contract = await network.getContract("invoicecc");

// try {
//   const list = await contract.addContractListener(
//     "InvoiceCreatedListener",
//     "InvoiceCreated",
//     (eror, event, blockNumber, transactionId, status) => {
//       if (err) {
//         console.log(err.message);
//         return;
//       }
//       console.log(
//         `Event:${event} BlockNumber:${blockNumber} TxID: ${transactionId} ${status}`
//       );
//     }
//   );
// } catch (err) {
//   console.log(err.message);
// }
// let eventhub = network
//   .getChannel("airlinechannel")
//   .getChannelEventHub("peer0.org1.example.com");

// let chaincodeHandler = await eventhub.registerChaincodeEvent(
//   "invoicecc",
//   "INVOICECREATED",
//   (event, blockNumber, txId, txStatus) => {
//     console.log(
//       `${event.chaincode_id} BN:${blockNumber} txID:${txId} ${txStatus}`
//     );
//   },
//   (err) => {
//     console.log(err.message);
//   }
// );

// eventhub.connect(true);
// console.log(chaincodeHandler);
