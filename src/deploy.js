const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();

  console.log("Deploying contracts with the account:", deployer.address);

  const CrossChainBridge = await hre.ethers.getContractFactory("CrossChainBridge");
  const bridge = await CrossChainBridge.deploy("<TOKEN_ADDRESS>");

  console.log("CrossChainBridge deployed to:", bridge.address);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

  