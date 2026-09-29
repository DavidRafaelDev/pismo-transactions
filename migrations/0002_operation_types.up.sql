CREATE TABLE operation_types (
    operation_type_id TINYINT UNSIGNED NOT NULL,
    description       VARCHAR(50) NOT NULL,
    PRIMARY KEY (operation_type_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
