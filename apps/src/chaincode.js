const { Gateway, FileSystemWallet } = require("fabric-network");
const fs = require("fs");
const yaml = require("js-yaml");

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
    identity: "joni@org1.example.com",
    discovery: { enabled: true },
  });
  console.log("connected");

  console.log("getNetwork");
  const network = await gateway.getNetwork("airlinechannel");
  const contract = await network.getContract("airplanecc");

  if (process.argv[2] === "query") {
    try {
      const res = await contract.evaluateTransaction(
        "QueryBySN",
        process.argv[3]
      );

      console.log(res.status);
      console.log(res.toString());
    } catch (err) {
      console.log(err);
    }
  } else if (process.argv[2] === "invoke") {
    try {
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
      console.log(res);
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
