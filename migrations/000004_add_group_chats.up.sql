ALTER TABLE chats ADD COLUMN chat_type TEXT NOT NULL DEFAULT 'direct';
ALTER TABLE chats ADD CONSTRAINT check_chat_type CHECK (chat_type IN ('direct', 'group'));
ALTER TABLE chats ADD COLUMN name TEXT; -- nullable, only meaningful for group chats
