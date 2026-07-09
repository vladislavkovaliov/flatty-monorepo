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