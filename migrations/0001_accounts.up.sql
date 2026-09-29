CREATE TABLE accounts (
    account_id      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    document_number VARCHAR(14) NOT NULL,
    created_at      DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (account_id),
    UNIQUE KEY uk_accounts_document_number (document_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
