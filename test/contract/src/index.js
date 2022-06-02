const fastify = require('fastify')({ logger: true });
const hre = require("hardhat");


// readiness check
fastify.get('/v1/healthz', async (req, reply) => {
    reply
        .code(200)
        .header('Content-Type', 'application/json; charset=utf-8')
        .send({ success: true })
});

fastify.get('/v1/deployContract', async (req, reply) => {
    const GLDToken = await hre.ethers.getContractFactory("GLDToken");
    const token = await GLDToken.deploy();
    await token.deployed();

    return {
        address: token.address,
        txHash: token.deployTransaction.hash,
        blockNumber: token.deployTransaction.blockNumber,
        blockHash: token.deployTransaction.blockHash,
    }
});

fastify.get('/v1/destroyContract', async (req, reply) => {
    const addr = req.query.addr;

    const Token = await hre.ethers.getContractFactory("GLDToken");
    const token = await Token.attach(addr);

    await token.destroy();
    const blockNum = await hre.ethers.provider.getBlockNumber()

    return {
        blockNumber: blockNum,
    }
})

fastify.get('/v1/sendEth', async (req, reply) => {
    const to = req.query.to;
    const value = req.query.value;

    const [owner] = await hre.ethers.getSigners();
    const tx = await owner.sendTransaction({
        to,
        value: hre.ethers.utils.parseEther(value)
    });
    await tx.wait(1)

    // console.log(tx);
    // const coinbaseBalance = await hre.ethers.provider.getBalance(owner.address);
    // const receiverBalance = await hre.ethers.provider.getBalance(to);
    // console.log(coinbaseBalance.toString(), receiverBalance.toString());

    return {
        from: tx.from,
        to: tx.to,
        //value: tx.value.toString(),
        txHash: tx.hash,
        blockNumber: tx.blockNumber,
        blockHash: tx.blockHash,
    }
});

fastify.get('/v1/deployTestContract', async (req, reply) => {
    const testContract = await hre.ethers.getContractFactory("Test");
    const test = await testContract.deploy();
    await test.deployed();

    return {
        address: test.address,
        txHash: test.deployTransaction.hash,
        blockNumber: test.deployTransaction.blockNumber,
        blockHash: test.deployTransaction.blockHash,
    }
});

fastify.get('/v1/putTestValue', async (req, reply) => {
    const addr = req.query.addr;
    const index = req.query.index;
    const value = req.query.value;

    const testContract = await hre.ethers.getContractFactory("Test");
    const test = await testContract.attach(addr);

    const tx = await test.Put(index, value);
    const receipt = await tx.wait();

    return {
        blockNumber: receipt.blockNumber,
    }
});

fastify.get('/v1/destroyTestContract', async (req, reply) => {
    const addr = req.query.addr;

    const testContract = await hre.ethers.getContractFactory("Test");
    const test = await testContract.attach(addr);

    await test.destroy();
    const blockNum = await hre.ethers.provider.getBlockNumber()

    return {
        blockNumber: blockNum,
    }
})

async function main() {
    try {
        await fastify.listen(3000, '0.0.0.0');
    } catch (err) {
        fastify.log.error(err);
        process.exit(1);
    }
}

main();
