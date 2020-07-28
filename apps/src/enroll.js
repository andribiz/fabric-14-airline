const {
  Wallet,
  FileSystemWallet,
  X509WalletMixin,
  Gateway,
} = require("fabric-network");
const BaseClient = require("fabric-client");
const jsrsa = require("jsrsasign");
const FabricCAService = require("fabric-ca-client");
const utils = require("fabric-ca-client");
const yaml = require("js-yaml");
const fs = require("fs");

const WALLET_PATH = "./wallet";

async function enrollAdmin(ca, wallet) {
  const enrollment = await ca.enroll({
    enrollmentID: "admin",
    enrollmentSecret: "adminpw",
  });

  //   console.log(enrollment);

  const identity = X509WalletMixin.createIdentity(
    "Org1MSP",
    enrollment.certificate,
    enrollment.key.toBytes()
  );

  await wallet.import("Admin@org1.example.com", identity);

  console.log("Enrollment Admin Done");
}

function generateCSR(privateKey, username) {
  const asn1 = jsrsa.asn1;
  const subjectDN = `CN=${username},C=ID,ST=JawaTimur,L=SBY,O=org1,OU=Admin`;
  const csr = asn1.csr.CSRUtil.newCSRPEM({
    sbjprvkey: privateKey.toBytes(),
    sbjpubkey: privateKey.getPublicKey().toBytes(),
    sigalg: "SHA256withECDSA",
    subject: { str: asn1.x509.X500Name.ldapToOneline(subjectDN) },
    ext: [
      {
        subjectAltName: {
          array: [{ dns: "org1.example.com" }, { dns: "localhost" }],
        },
      },
    ],
  });
  // const csr = privateKey.generateCSR(subjectDN, ext);
  return csr;
}

async function createUser(ca, wallet, adminID, username, password) {
  const idRegister = {
    enrollmentID: username,
    enrollmentSecret: password,
    role: "client",
    affiliation: "org1",
    maxEnrollments: 10,
    attrs: [{ name: "apps.admin", value: "false", ecert: true }],
  };
  const reg = await ca.register(idRegister, adminID);

  const privateKey = await BaseClient.newCryptoSuite().generateKey({
    ephemeral: true,
  });

  const csr = generateCSR(privateKey, username);

  console.log(csr);
  const enrollment = await ca.enroll({
    enrollmentID: username,
    enrollmentSecret: password,
    csr: csr,
  });
  console.log(enrollment);

  const identity = X509WalletMixin.createIdentity(
    "Org1MSP",
    enrollment.certificate,
    // enrollment.key.toBytes()
    privateKey.toBytes()
  );
  await wallet.import(`${username}@org1.example.com`, identity);
  console.log("Register And Enrollemt complete");
}

async function enrollAgain(ca, wallet, username, password) {
  const privateKey = await BaseClient.newCryptoSuite().generateKey({
    ephemeral: true,
  });

  const csr = generateCSR(privateKey, username);

  console.log(csr);
  const enrollment = await ca.enroll({
    enrollmentID: username,
    enrollmentSecret: password,
    csr: csr,
  });
  console.log(enrollment);

  const identity = X509WalletMixin.createIdentity(
    "Org1MSP",
    enrollment.certificate,
    // enrollment.key.toBytes()
    privateKey.toBytes()
  );
  await wallet.import(`${username}@org1.example.com`, identity);
  console.log("Enroll Again complete");
}

async function reEnroll(ca, wallet, user) {
  const privateKey = await BaseClient.newCryptoSuite().generateKey({
    ephemeral: true,
  });

  const enrollment = await ca.reenroll(user);

  const identity = X509WalletMixin.createIdentity(
    "Org1MSP",
    enrollment.certificate,
    enrollment.key.toBytes()
  );

  await wallet.import(`${user.getName()}@org1.example.com`, identity);
  console.log("Register And Enrollemt complete");
}

async function main() {
  let conn = yaml.safeLoad(
    fs.readFileSync(
      "../../network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml",
      "utf8"
    )
  );

  const caInfo = conn.certificateAuthorities["ca.org1.com"];
  const ca = new FabricCAService(
    caInfo.url,
    { trustedRoots: caInfo.tlsCACerts.pem, verify: false },
    caInfo.name
  );
  const wallet = new FileSystemWallet(WALLET_PATH);

  console.log(process.argv[2]);
  if (process.argv[2] === "enrollAdmin") {
    await enrollAdmin(ca, wallet);
  } else if (process.argv[2] === "createUser") {
    console.log("Creating User");
    const gateway = new Gateway();
    await gateway.connect(conn, {
      wallet,
      identity: "Admin@org1.example.com",
      discovery: { enabled: true },
    });
    const adminID = gateway.getCurrentIdentity();
    await createUser(ca, wallet, adminID, process.argv[3], process.argv[4]);
  } else if (process.argv[2] === "reenroll") {
    console.log("Reenroll User");
    const gateway = new Gateway();
    await gateway.connect(conn, {
      wallet,
      identity: `${process.argv[3]}@org1.example.com`,
      discovery: { enabled: true },
    });
    const userID = gateway.getCurrentIdentity();
    await reEnroll(ca, wallet, userID);
  } else if (process.argv[2] === "enrollAgain") {
    await enrollAgain(ca, wallet, process.argv[3], process.argv[4]);
  } else if (process.argv[2] === "data") {
    console.log(await wallet.export("Admin@org1.example.com"));
  }
}

try {
  main();
} catch (err) {
  console.error(err);
}
