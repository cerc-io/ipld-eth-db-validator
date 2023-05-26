const fastify = require('fastify')({ logger: true });
const hre = require("hardhat");

// readiness check
fastify.get('/v1/healthz', async (req, reply) => {
  reply
    .code(200)
    .header('Content-Type', 'application/json; charset=utf-8')
    .send({ success: true })
});

fastify.get('/v1/sendEth', async (req, reply) => {
  const to = req.query.to;
  const value = hre.ethers.utils.parseEther(req.query.value);

  const owner = await hre.ethers.getSigner();
  const tx = await owner.sendTransaction({to, value}).then(tx => tx.wait());

  return {
    from: tx.from,
    to: tx.to,
    txHash: tx.hash,
    blockNumber: tx.blockNumber,
    blockHash: tx.blockHash,
  }
});

function contractCreator(name) {
  return async (req, reply) => {
    const contract = await hre.ethers.getContractFactory(name);
    const instance = await contract.deploy();
    const rct = await instance.deployTransaction.wait();

    return {
      address: instance.address,
      txHash: rct.transactionHash,
      blockNumber: rct.blockNumber,
      blockHash: rct.blockHash,
    }
  }
}

function contractDestroyer(name) {
  return async (req, reply) => {
    const addr = req.query.addr;
    const contract = await hre.ethers.getContractFactory(name);
    const instance = contract.attach(addr);
    const rct = await instance.destroy().then(tx => tx.wait());

    return {
      blockNumber: rct.blockNumber,
    }
  }
}

fastify.get('/v1/deployContract', contractCreator("GLDToken"));
fastify.get('/v1/destroyContract', contractDestroyer("GLDToken"));

fastify.get('/v1/deployTestContract', contractCreator("Test"));
fastify.get('/v1/destroyTestContract', contractDestroyer("Test"));

fastify.get('/v1/putTestValue', async (req, reply) => {
  const addr = req.query.addr;
  const value = req.query.value;

  const testContract = await hre.ethers.getContractFactory("Test");
  const test = await testContract.attach(addr);

  const rct = await test.Put(value).then(tx => tx.wait());

  return {
    blockNumber: rct.blockNumber,
  }
});

async function main() {
  try {
    await fastify.listen({ port: 3000, host: '0.0.0.0' });
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
}

process.on('SIGINT', () => fastify.close().then(() => process.exit(1)));

main();
