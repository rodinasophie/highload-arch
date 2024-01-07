\c dialogs_social_net;
SET ROLE TO admin_user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS dialogs (
    id UUID DEFAULT uuid_generate_v4(),
    author_id UUID NOT NULL, 
    recepient_id UUID NOT NULL,
    dialog_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    text VARCHAR(400) NOT NULL,
    PRIMARY KEY(id, dialog_id) 
);