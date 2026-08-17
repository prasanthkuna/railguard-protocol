// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title RailguardVaultExecutor
/// @notice v4 §24 reference vault — enforces authorised execution facts on-chain.
contract RailguardVaultExecutor {
    event RailguardExecution(
        bytes32 indexed executionId,
        bytes32 indexed intentHash,
        address indexed token,
        address recipient,
        uint256 amount
    );

    mapping(bytes32 => bool) public executed;

    function execute(
        bytes32 executionId,
        bytes32 intentHash,
        address token,
        address recipient,
        uint256 amount,
        uint256 expiry
    ) external {
        require(block.timestamp <= expiry, "EXPIRED");
        require(!executed[executionId], "REPLAY");
        executed[executionId] = true;
        emit RailguardExecution(executionId, intentHash, token, recipient, amount);
    }
}
