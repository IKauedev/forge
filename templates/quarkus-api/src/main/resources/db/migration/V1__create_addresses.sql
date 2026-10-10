CREATE TABLE addresses (
    cep          VARCHAR(8)   PRIMARY KEY,
    street       VARCHAR(255),
    complement   VARCHAR(255),
    neighborhood VARCHAR(255),
    city         VARCHAR(255),
    state        VARCHAR(2)
);
