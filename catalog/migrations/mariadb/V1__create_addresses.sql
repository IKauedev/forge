CREATE TABLE addresses (
    cep          VARCHAR(8)   NOT NULL,
    street       VARCHAR(255),
    complement   VARCHAR(255),
    neighborhood VARCHAR(255),
    city         VARCHAR(255),
    state        VARCHAR(2),
    PRIMARY KEY (cep)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
