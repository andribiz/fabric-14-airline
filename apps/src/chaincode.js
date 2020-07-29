const { Gateway, FileSystemWallet } = require("fabric-network");
const fs = require("fs");
const yaml = require("js-yaml");
const { networkInterfaces } = require("os");

const WALLET_PATH = "./wallet";
const CCP_PATH =
  "../../network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml";

async function main() {
  const ccp = yaml.safeLoad(fs.readFileSync(CCP_PATH, "utf8"));

  const gateway = new Gateway();
  const wallet = new FileSystemWallet(WALLET_PATH);

  console.log("connecting.....");
  await gateway.connect(ccp, {
    wallet,
    identity: "john@org1.example.com",
    discovery: { enabled: true },
  });
  console.log("connected");

  console.log("getNetwork");
  const network = await gateway.getNetwork("airlinechannel");

  if (process.argv[2] === "query") {
    try {
      const contract = await network.getContract("airplanecc");
      const res = await contract.evaluateTransaction(
        "QueryBySN",
        process.argv[3]
      );

      console.log(res.status);
      console.log(res.toString());
    } catch (err) {
      console.log(err);
    }
  } else if (process.argv[2] === "invokeAirplane") {
    try {
      const contract = await network.getContract("airplanecc");

      const res = await contract.submitTransaction(
        "CreatePlane",
        process.argv[3],
        "Batch1",
        "2012-01-02",
        "BA-737MAx",
        "2",
        "123023,123123",
        "200",
        "1000.5",
        "900.5"
      );
      console.log(res.toString());
    } catch (err) {
      console.log(err.message);
    }
  } else if (process.argv[2] === "invokeInvoice") {
    try {
      const contract = await network.getContract("invoicecc");

      const res = await contract.submitTransaction(
        "CreateInvoice",
        process.argv[3],
        process.argv[4],
        process.argv[5],
        process.argv[6],
        process.argv[7],
        process.argv[8]
      );
      console.log(res.toString());
    } catch (err) {
      console.log(err.message);
    }
  } else if (process.argv[2] === "queryByKeys") {
    try {
      const contract = await network.getContract("invoicecc");
      const res = await contract.evaluateTransaction(
        "GetInvoiceByPartner",
        process.argv[3]
      );
      console.log(res.toString());
    } catch (err) {
      console.log(err.message);
    }
  } else if (process.argv[2] === "queryByPartner") {
    try {
      const contract = await network.getContract("invoicecc");
      const res = await contract.evaluateTransaction(
        "QueryInvoiceByPartner",
        process.argv[3],
        ""
      );
      console.log(res.toString());
    } catch (err) {
      console.log(err.message);
    }
  } else if (process.argv[2] === "createInvoiceWithTransient") {
    invlines = [
      { Product: "Produc1", ProductQty: 3, ProductUom: "Pcs", Price: 10000 },
      { Product: "Produc2", ProductQty: 6, ProductUom: "Pcs", Price: 30000 },
    ];

    try {
      const contract = await network.getContract("invoicecc");
      const tx = contract.createTransaction("CreateInvoice");
      const res = await tx
        .setTransient({
          InvoiceLines: Buffer.from(JSON.stringify(invlines)),
        })
        .submit(
          process.argv[3],
          process.argv[4],
          process.argv[5],
          process.argv[6],
          process.argv[7],
          process.argv[8]
        );
      console.log(res.toString());
    } catch (err) {
      console.log(err.message);
    }
  } else if (process.argv[2] === "getInvoiceLines") {
    try {
      const contract = await network.getContract("invoicecc");
      const res = await contract.evaluateTransaction(
        "GetInvoiceLines",
        process.argv[3],
        ""
      );
      console.log(res.toString());
    } catch (err) {
      console.log(err.message);
    }
  }

  gateway.disconnect();
}

try {
  main();
} catch (err) {
  console.err(err);
}
