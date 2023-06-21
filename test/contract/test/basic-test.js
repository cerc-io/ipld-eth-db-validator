import { expect } from 'chai';

describe("GLDToken", function() {
  it("Should return the owner's balance", async function() {
    const Token = await ethers.getContractFactory("GLDToken");
    const token = await Token.deploy();
    await token.deployed();
    
    const [owner] = await ethers.getSigners();
    const balance = await token.balanceOf(owner.address);
    expect(balance).to.equal('1000000000000000000000');
    expect(await token.totalSupply()).to.equal(balance);  });
});
