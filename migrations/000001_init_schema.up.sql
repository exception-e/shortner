create table if not exists links (
    id BIGSERIAL PRIMARY KEY,
    short_link TEXT UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

create index idx_short_link on links (short_link);
create index idx_original_url on links (original_url);
