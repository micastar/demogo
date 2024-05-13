CREATE TABLE IF NOT EXISTS rss_post (
    g_id TEXT NOT NULL PRIMARY KEY, -- guid
    title TEXT NOT NULL,
    descrp TEXT, -- description
    logdate TEXT NOT NULL, -- published
    link TEXT NOT NULL,
    cates TEXT[] -- categories
);

CREATE TABLE contents (
    g_id TEXT NOT NULL, -- guid
    title TEXT NOT NULL,
    descrp TEXT, -- description
    logdate TEXT NOT NULL, -- published
    link TEXT NOT NULL,
    cates TEXT[], -- categories
    PRIMARY KEY (g_id,logdate)
)  PARTITION BY RANGE (logdate);

insert into contents (id, title, descrp, published, link, categories) 
values(
    'news634380137',
    'EU queries X over cut to content moderation resources', 
    'The EU on .', 
    '2024-05-08', 
    'https://techxplore.com/news/2024-05-eu-queries-content-moderation-resources.html', 
    '{Business}'
);