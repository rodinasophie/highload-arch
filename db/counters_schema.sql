\c counters_social_net;
SET ROLE TO admin_user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS unread_messages (
    id UUID DEFAULT uuid_generate_v4(),
    author_id UUID NOT NULL, 
    recepient_id UUID NOT NULL,
    count INTEGER DEFAULT 1 NOT NULL,
    PRIMARY KEY(author_id, recepient_id) 
);