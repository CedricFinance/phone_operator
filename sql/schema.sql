-- CREATE USER `phone_operator`@`%` IDENTIFIED BY 'LCJPRWDXVDRhaKxWzRgdjGnjUz9KQWVW';
-- CREATE SCHEMA phone_operator;
-- GRANT ALL PRIVILEGES ON `phone_operator`.* TO `phone_operator`@`%`; 

CREATE TABLE ForwardingRequests(
    id CHAR(36) PRIMARY KEY,
    requester_id VARCHAR(16) NOT NULL,
    requester_name VARCHAR(50) NOT NULL,
    duration INT NOT NULL,
    created_at DATETIME(3) NOT NULL,
    accepted_at DATETIME(3),
    refused_at DATETIME(3),
    expires_at DATETIME(3),
    answered_by VARCHAR(16)
) CHARACTER SET utf8mb4;
