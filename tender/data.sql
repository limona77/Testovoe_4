CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE employee (
                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                          username VARCHAR(50) UNIQUE NOT NULL,
                          first_name VARCHAR(50),
                          last_name VARCHAR(50),
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
);

CREATE TABLE organization (
                              id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                              name VARCHAR(100) NOT NULL,
                              description TEXT,
                              type organization_type,
                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                              updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible (
                                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                          organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
                                          user_id UUID REFERENCES employee(id) ON DELETE CASCADE
);


CREATE TABLE tender (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
                        title VARCHAR(255) NOT NULL,
                        description TEXT,
                        service_type VARCHAR(100),
                        status VARCHAR(20) NOT NULL,
                        version INT DEFAULT 1,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        creator_username VARCHAR(50)
);
CREATE TABLE bids (
                      id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                      tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
                      organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
                      title VARCHAR(255) NOT NULL,
                      description TEXT,
                      status VARCHAR(20) NOT NULL,
                      version INT DEFAULT 1,
                      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                      creator_username VARCHAR(50)
);

ALTER TABLE bids
    ADD COLUMN decision VARCHAR(20);