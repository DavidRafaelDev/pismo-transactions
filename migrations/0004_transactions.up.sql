CREATE TABLE transactions (
    transaction_id     BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    account_id         BIGINT UNSIGNED NOT NULL,
    operation_type_id  TINYINT UNSIGNED NOT NULL,
    amount             DECIMAL(15,2) NOT NULL,
    event_date         DATETIME(6) NOT NULL,
    PRIMARY KEY (transaction_id),
    KEY idx_transactions_account (account_id),
    CONSTRAINT fk_transactions_account
        FOREIGN KEY (account_id) REFERENCES accounts(account_id),
    CONSTRAINT fk_transactions_operation_type
        FOREIGN KEY (operation_type_id) REFERENCES operation_types(operation_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
