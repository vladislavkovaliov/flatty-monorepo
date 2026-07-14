# create resident table 

CREATE TABLE resident_locations (
    id BIGSERIAL PRIMARY KEY,

    country VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    street VARCHAR(150) NOT NULL,
    house VARCHAR(20) NOT NULL,
    apartment VARCHAR(20),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


# insert fake data into table

INSERT INTO resident_locations (
    country,
    city,
    postal_code,
    street,
    house,
    apartment
) VALUES (
    'Belarus',
    'Minsk',
    '220045',
    'Blurish Red',
    '16',
    '25'
);

# create trigger to update updated_at when record is updated

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER set_updated_at
BEFORE UPDATE ON resident_locations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


CREATE TABLE categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE expenses (
    id                   BIGSERIAL PRIMARY KEY,
    resident_location_id BIGINT NOT NULL REFERENCES resident_locations(id) ON DELETE CASCADE,
    category_id          BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount               NUMERIC(12, 2) NOT NULL,
    month                INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    year                 INTEGER NOT NULL CHECK (year >= 2000),
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (resident_location_id, category_id, month, year)
);

INSERT INTO categories (name, description) VALUES
('utilities',       'Коммунальные платежи'),
('additional',      'Дополнительные платежи'),
('electricity',     'Электроэнергия'),
('internet_tv',     'ZALA, byfly, Умный дом, пакеты (A1)'),
('mts_1',           'МТС (292915789)'),
('mts_2',           'МТС (336615860)'),
('parking_milya',   'Парковка Михаила'),
('parking_kamova',  'Парковка Камова д5');


-- April 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 101.63, 4, 2023),
(1, 2, 0.00, 4, 2023),
(1, 3, 0.00, 4, 2023),
(1, 4, 0.00, 4, 2023),
(1, 5, 0.00, 4, 2023),
(1, 6, 0.00, 4, 2023),
(1, 7, 0.00, 4, 2023),
(1, 8, 0.00, 4, 2023);

-- May 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 10.32, 5, 2023),
(1, 2, 10.46, 5, 2023),
(1, 3, 0.00, 5, 2023),
(1, 4, 0.00, 5, 2023),
(1, 5, 0.00, 5, 2023),
(1, 6, 0.00, 5, 2023),
(1, 7, 0.00, 5, 2023),
(1, 8, 0.00, 5, 2023);

-- June 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 0.00, 6, 2023),
(1, 2, 10.46, 6, 2023),
(1, 3, 0.00, 6, 2023),
(1, 4, 0.00, 6, 2023),
(1, 5, 0.00, 6, 2023),
(1, 6, 0.00, 6, 2023),
(1, 7, 0.00, 6, 2023),
(1, 8, 0.00, 6, 2023);

-- July 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 33.84, 7, 2023),
(1, 2, 10.46, 7, 2023),
(1, 3, 0.00, 7, 2023),
(1, 4, 0.00, 7, 2023),
(1, 5, 0.00, 7, 2023),
(1, 6, 0.00, 7, 2023),
(1, 7, 0.00, 7, 2023),
(1, 8, 0.00, 7, 2023);

-- August 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 33.85, 8, 2023),
(1, 2, 11.64, 8, 2023),
(1, 3, 0.00, 8, 2023),
(1, 4, 0.00, 8, 2023),
(1, 5, 0.00, 8, 2023),
(1, 6, 0.00, 8, 2023),
(1, 7, 0.00, 8, 2023),
(1, 8, 0.00, 8, 2023);

-- September 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 33.87, 9, 2023),
(1, 2, 11.51, 9, 2023),
(1, 3, 72.11, 9, 2023),
(1, 4, 25.00, 9, 2023),
(1, 5, 20.00, 9, 2023),
(1, 6, 0.00, 9, 2023),
(1, 7, 0.00, 9, 2023),
(1, 8, 0.00, 9, 2023);

-- October 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 34.57, 10, 2023),
(1, 2, 11.51, 10, 2023),
(1, 3, 0.00, 10, 2023),
(1, 4, 20.00, 10, 2023),
(1, 5, 0.00, 10, 2023),
(1, 6, 0.00, 10, 2023),
(1, 7, 0.00, 10, 2023),
(1, 8, 0.00, 10, 2023);

-- November 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 78.37, 11, 2023),
(1, 2, 12.29, 11, 2023),
(1, 3, 0.00, 11, 2023),
(1, 4, 0.00, 11, 2023),
(1, 5, 0.00, 11, 2023),
(1, 6, 0.00, 11, 2023),
(1, 7, 0.00, 11, 2023),
(1, 8, 0.00, 11, 2023);

-- December 2023
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 115.97, 12, 2023),
(1, 2, 11.71, 12, 2023),
(1, 3, 23.62, 12, 2023),
(1, 4, 25.00, 12, 2023),
(1, 5, 0.00, 12, 2023),
(1, 6, 0.00, 12, 2023),
(1, 7, 102.75, 12, 2023),
(1, 8, 0.00, 12, 2023);

-- January 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 124.09, 1, 2024),
(1, 2, 12.70, 1, 2024),
(1, 3, 25.08, 1, 2024),
(1, 4, 0.00, 1, 2024),
(1, 5, 15.00, 1, 2024),
(1, 6, 17.90, 1, 2024),
(1, 7, 32.69, 1, 2024),
(1, 8, 0.00, 1, 2024);

-- February 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 162.98, 2, 2024),
(1, 2, 11.71, 2, 2024),
(1, 3, 30.40, 2, 2024),
(1, 4, 20.45, 2, 2024),
(1, 5, 15.00, 2, 2024),
(1, 6, 17.90, 2, 2024),
(1, 7, 28.78, 2, 2024),
(1, 8, 0.00, 2, 2024);

-- March 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 129.48, 3, 2024),
(1, 2, 11.71, 3, 2024),
(1, 3, 29.54, 3, 2024),
(1, 4, 20.45, 3, 2024),
(1, 5, 15.00, 3, 2024),
(1, 6, 17.90, 3, 2024),
(1, 7, 28.78, 3, 2024),
(1, 8, 0.00, 3, 2024);

-- April 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 127.64, 4, 2024),
(1, 2, 11.71, 4, 2024),
(1, 3, 36.44, 4, 2024),
(1, 4, 0.00, 4, 2024),
(1, 5, 0.00, 4, 2024),
(1, 6, 62.00, 4, 2024),
(1, 7, 28.78, 4, 2024),
(1, 8, 0.00, 4, 2024);

-- May 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 86.67, 5, 2024),
(1, 2, 12.58, 5, 2024),
(1, 3, 30.18, 5, 2024),
(1, 4, 20.45, 5, 2024),
(1, 5, 15.00, 5, 2024),
(1, 6, 22.16, 5, 2024),
(1, 7, 59.81, 5, 2024),
(1, 8, 0.00, 5, 2024);

-- June 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 39.32, 6, 2024),
(1, 2, 12.58, 6, 2024),
(1, 3, 30.62, 6, 2024),
(1, 4, 30.00, 6, 2024),
(1, 5, 25.00, 6, 2024),
(1, 6, 20.00, 6, 2024),
(1, 7, 32.43, 6, 2024),
(1, 8, 0.00, 6, 2024);

-- July 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 39.33, 7, 2024),
(1, 2, 12.58, 7, 2024),
(1, 3, 51.53, 7, 2024),
(1, 4, 30.00, 7, 2024),
(1, 5, 25.00, 7, 2024),
(1, 6, 20.00, 7, 2024),
(1, 7, 31.17, 7, 2024),
(1, 8, 0.00, 7, 2024);

-- August 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 39.33, 8, 2024),
(1, 2, 12.58, 8, 2024),
(1, 3, 65.54, 8, 2024),
(1, 4, 30.00, 8, 2024),
(1, 5, 20.00, 8, 2024),
(1, 6, 20.00, 8, 2024),
(1, 7, 31.58, 8, 2024),
(1, 8, 0.00, 8, 2024);

-- September 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 64.03, 9, 2024),
(1, 2, 12.58, 9, 2024),
(1, 3, 39.34, 9, 2024),
(1, 4, 31.13, 9, 2024),
(1, 5, 25.00, 9, 2024),
(1, 6, 20.00, 9, 2024),
(1, 7, 31.52, 9, 2024),
(1, 8, 0.00, 9, 2024);

-- October 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 332.80, 10, 2024),
(1, 2, 12.58, 10, 2024),
(1, 3, 51.53, 10, 2024),
(1, 4, 30.00, 10, 2024),
(1, 5, 25.00, 10, 2024),
(1, 6, 20.00, 10, 2024),
(1, 7, 31.24, 10, 2024),
(1, 8, 0.00, 10, 2024);

-- November 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 90.24, 11, 2024),
(1, 2, 12.58, 11, 2024),
(1, 3, 53.90, 11, 2024),
(1, 4, 25.00, 11, 2024),
(1, 5, 25.00, 11, 2024),
(1, 6, 20.00, 11, 2024),
(1, 7, 32.93, 11, 2024),
(1, 8, 0.00, 11, 2024);

-- December 2024
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 149.07, 12, 2024),
(1, 2, 12.58, 12, 2024),
(1, 3, 66.79, 12, 2024),
(1, 4, 24.00, 12, 2024),
(1, 5, 25.00, 12, 2024),
(1, 6, 20.00, 12, 2024),
(1, 7, 34.85, 12, 2024),
(1, 8, 0.00, 12, 2024);

-- January 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 147.34, 1, 2025),
(1, 2, 12.69, 1, 2025),
(1, 3, 35.36, 1, 2025),
(1, 4, 19.00, 1, 2025),
(1, 5, 25.00, 1, 2025),
(1, 6, 20.00, 1, 2025),
(1, 7, 37.62, 1, 2025),
(1, 8, 0.00, 1, 2025);

-- February 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 152.05, 2, 2025),
(1, 2, 12.58, 2, 2025),
(1, 3, 47.76, 2, 2025),
(1, 4, 19.00, 2, 2025),
(1, 5, 35.00, 2, 2025),
(1, 6, 20.00, 2, 2025),
(1, 7, 40.24, 2, 2025),
(1, 8, 0.00, 2, 2025);

-- March 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 146.02, 3, 2025),
(1, 2, 31.65, 3, 2025),
(1, 3, 54.03, 3, 2025),
(1, 4, 35.00, 3, 2025),
(1, 5, 25.00, 3, 2025),
(1, 6, 20.00, 3, 2025),
(1, 7, 41.06, 3, 2025),
(1, 8, 0.00, 3, 2025);

-- April 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 158.26, 4, 2025),
(1, 2, 50.67, 4, 2025),
(1, 3, 89.24, 4, 2025),
(1, 4, 19.00, 4, 2025),
(1, 5, 25.00, 4, 2025),
(1, 6, 20.00, 4, 2025),
(1, 7, 38.26, 4, 2025),
(1, 8, 0.00, 4, 2025);

-- May 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 110.51, 5, 2025),
(1, 2, 50.92, 5, 2025),
(1, 3, 108.06, 5, 2025),
(1, 4, 38.00, 5, 2025),
(1, 5, 25.00, 5, 2025),
(1, 6, 20.00, 5, 2025),
(1, 7, 38.26, 5, 2025),
(1, 8, 37.18, 5, 2025);

-- June 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 194.64, 6, 2025),
(1, 2, 50.92, 6, 2025),
(1, 3, 44.38, 6, 2025),
(1, 4, 19.00, 6, 2025),
(1, 5, 25.00, 6, 2025),
(1, 6, 0.00, 6, 2025),
(1, 7, 29.21, 6, 2025),
(1, 8, 56.16, 6, 2025);

-- July 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 66.77, 7, 2025),
(1, 2, 50.92, 7, 2025),
(1, 3, 61.75, 7, 2025),
(1, 4, 19.00, 7, 2025),
(1, 5, 25.00, 7, 2025),
(1, 6, 0.00, 7, 2025),
(1, 7, 29.21, 7, 2025),
(1, 8, 37.75, 7, 2025);

-- August 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 74.92, 8, 2025),
(1, 2, 50.92, 8, 2025),
(1, 3, 80.80, 8, 2025),
(1, 4, 19.00, 8, 2025),
(1, 5, 25.00, 8, 2025),
(1, 6, 0.00, 8, 2025),
(1, 7, 29.91, 8, 2025),
(1, 8, 0.00, 8, 2025);

-- September 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 69.19, 9, 2025),
(1, 2, 50.92, 9, 2025),
(1, 3, 69.71, 9, 2025),
(1, 4, 19.00, 9, 2025),
(1, 5, 25.00, 9, 2025),
(1, 6, 0.00, 9, 2025),
(1, 7, 29.21, 9, 2025),
(1, 8, 0.00, 9, 2025);

-- October 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 65.69, 10, 2025),
(1, 2, 50.92, 10, 2025),
(1, 3, 85.87, 10, 2025),
(1, 4, 19.00, 10, 2025),
(1, 5, 25.00, 10, 2025),
(1, 6, 0.00, 10, 2025),
(1, 7, 29.21, 10, 2025),
(1, 8, 0.00, 10, 2025);

-- November 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 91.81, 11, 2025),
(1, 2, 50.92, 11, 2025),
(1, 3, 62.71, 11, 2025),
(1, 4, 19.00, 11, 2025),
(1, 5, 25.00, 11, 2025),
(1, 6, 0.00, 11, 2025),
(1, 7, 31.61, 11, 2025),
(1, 8, 0.00, 11, 2025);

-- December 2025
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 176.45, 12, 2025),
(1, 2, 50.92, 12, 2025),
(1, 3, 60.54, 12, 2025),
(1, 4, 19.00, 12, 2025),
(1, 5, 25.00, 12, 2025),
(1, 6, 0.00, 12, 2025),
(1, 7, 245.97, 12, 2025),
(1, 8, 0.00, 12, 2025);

-- January 2026
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 128.98, 1, 2026),
(1, 2, 50.92, 1, 2026),
(1, 3, 50.41, 1, 2026),
(1, 4, 19.00, 1, 2026),
(1, 5, 25.00, 1, 2026),
(1, 6, 0.00, 1, 2026),
(1, 7, 33.75, 1, 2026),
(1, 8, 0.00, 1, 2026);

-- February 2026
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 226.89, 2, 2026),
(1, 2, 50.92, 2, 2026),
(1, 3, 63.68, 2, 2026),
(1, 4, 20.00, 2, 2026),
(1, 5, 25.00, 2, 2026),
(1, 6, 0.00, 2, 2026),
(1, 7, 34.89, 2, 2026),
(1, 8, 0.00, 2, 2026);

-- March 2026
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 185.43, 3, 2026),
(1, 2, 50.86, 3, 2026),
(1, 3, 71.88, 3, 2026),
(1, 4, 30.00, 3, 2026),
(1, 5, 25.00, 3, 2026),
(1, 6, 0.00, 3, 2026),
(1, 7, 34.92, 3, 2026),
(1, 8, 0.00, 3, 2026);

-- April 2026
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 144.60, 4, 2026),
(1, 2, 50.86, 4, 2026),
(1, 3, 90.34, 4, 2026),
(1, 4, 27.00, 4, 2026),
(1, 5, 25.00, 4, 2026),
(1, 6, 0.00, 4, 2026),
(1, 7, 34.69, 4, 2026),
(1, 8, 0.00, 4, 2026);

-- May 2026
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 147.18, 5, 2026),
(1, 2, 51.75, 5, 2026),
(1, 3, 98.46, 5, 2026),
(1, 4, 19.00, 5, 2026),
(1, 5, 0.00, 5, 2026),
(1, 6, 0.00, 5, 2026),
(1, 7, 33.16, 5, 2026),
(1, 8, 0.00, 5, 2026);

-- June 2026
INSERT INTO expenses (resident_location_id, category_id, amount, month, year) VALUES
(1, 1, 115.62, 6, 2026),
(1, 2, 51.85, 6, 2026),
(1, 3, 79.22, 6, 2026),
(1, 4, 19.00, 6, 2026),
(1, 5, 25.00, 6, 2026),
(1, 6, 0.00, 6, 2026),
(1, 7, 31.88, 6, 2026),
(1, 8, 0.00, 6, 2026);

-- expense stats materialized tables

CREATE TABLE expense_monthly_totals (
    month       INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    year        INTEGER NOT NULL CHECK (year >= 2000),
    total_spent NUMERIC(12,2) NOT NULL DEFAULT 0,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (month, year)
);

CREATE TABLE expense_monthly_averages (
    month          INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    year           INTEGER NOT NULL CHECK (year >= 2000),
    average_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    expense_count  INTEGER NOT NULL DEFAULT 0,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (month, year)
);

CREATE TRIGGER set_expense_monthly_totals_updated_at
BEFORE UPDATE ON expense_monthly_totals
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_expense_monthly_averages_updated_at
BEFORE UPDATE ON expense_monthly_averages
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();