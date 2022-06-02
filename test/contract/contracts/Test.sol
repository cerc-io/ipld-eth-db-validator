// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.0;

contract Test {
  address payable owner;

  modifier onlyOwner {
    require(
      msg.sender == owner,
      "Only owner can call this function."
    );
    _;
  }

  uint256[100] data;

  constructor() {
    owner = payable(msg.sender);
    data = [1];
  }

  function Put(uint256 addr, uint256 value) public {
    data[addr] = value;
  }

  function close() public onlyOwner {
    selfdestruct(owner);
  }
}
