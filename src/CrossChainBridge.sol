// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract CrossChainBridge is Ownable {
    ERC20 public token;
    mapping(address => uint256) public lockedTokens;
    event TokensLocked(address indexed user, uint256 amount);
    event TokensMinted(address indexed user, uint256 amount);
    event TokensUnlocked(address indexed user, uint256 amount);

    constructor(address _tokenAddress) {
        token = ERC20(_tokenAddress);
    }

    // Lock tokens on the source chain
    function lockTokens(uint256 _amount) external {
        require(_amount > 0, "Amount must be greater than zero");
        require(token.balanceOf(msg.sender) >= _amount, "Insufficient balance");

        // Transfer tokens to the contract
        token.transferFrom(msg.sender, address(this), _amount);
        lockedTokens[msg.sender] += _amount;

        emit TokensLocked(msg.sender, _amount);
    }

    // Mint tokens on the destination chain
    function mintTokens(address _to, uint256 _amount) external onlyOwner {
        require(_amount > 0, "Amount must be greater than zero");

        // Simulate minting (Can connect to another chain for real minting)
        token.transfer(_to, _amount);

        emit TokensMinted(_to, _amount);
    }

    // Unlock tokens on the source chain
    function unlockTokens(uint256 _amount) external {
        require(_amount > 0, "Amount must be greater than zero");
        require(lockedTokens[msg.sender] >= _amount, "Insufficient locked tokens");

        lockedTokens[msg.sender] -= _amount;

        // Transfer tokens back to the user
        token.transfer(msg.sender, _amount);

        emit TokensUnlocked(msg.sender, _amount);
    }
}
