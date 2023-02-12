// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.0;

contract Test {
    event log_put (address, uint256);

    address payable owner;
    mapping(address => uint256) public data;

    modifier onlyOwner {
        require(msg.sender == owner, "Only owner can call this function.");
        _;
    }

    constructor() {
        owner = payable(msg.sender);
    }

    function Put(uint256 value) public {
        emit log_put(msg.sender, value);

        data[msg.sender] = value;
    }

    function close() public onlyOwner {
        owner.transfer(address(this).balance);
    }
}
