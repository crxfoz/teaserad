CREATE TABLE hits
(
    view_id     UUID,
    banner_id   UInt64,
    platform_id UInt64,
    user_agent  String,
    device      String,
    created_at  UInt64,
    day         Date,
    dt          DateTime
) ENGINE = MergeTree() PARTITION BY toYYYYMM(dt)
      ORDER BY
          (banner_id, platform_id, dt);

CREATE TABLE hits_daily
(
    banner_id   UInt64,
    platform_id UInt64,
    hits        UInt64,
    day         Date
) ENGINE = SummingMergeTree(day,
           (day,
            banner_id,
            platform_id),
           8192);

CREATE TABLE kafka_hits
(
    banner_id   UInt64,
    platform_id UInt64,
    user_agent  String,
    device      String,
    created_at  UInt64
) ENGINE = Kafka('kafka-1:9092,kafka-2:9092,kafka-3:9092',
           'adshow.action.show',
           'ch-stat-hits',
           'JSONEachRow');

CREATE MATERIALIZED VIEW consumer_hits TO hits AS
SELECT banner_id,
       platform_id,
       user_agent,
       device,
       created_at,
       toDate(
               toDateTime(created_at)) AS day,
       toDateTime(
               created_at)             AS dt
FROM kafka_hits;

CREATE MATERIALIZED VIEW hits_daily_view TO hits_daily AS
SELECT banner_id,
       platform_id,
       count(
           ) AS hits,
       day
FROM hits
GROUP BY (
          day,
          banner_id,
          platform_id
             );

CREATE TABLE clicks
(
    banner_id   UInt64,
    platform_id UInt64,
    price       Float64,
    view_id     String,
    created_at  UInt64,
    day         Date,
    dt          DateTime
) ENGINE = MergeTree() PARTITION BY toYYYYMM(dt)
      ORDER BY
          (banner_id, platform_id, dt);

CREATE TABLE clicks_daily
(
    banner_id   UInt64,
    platform_id UInt64,
    clicks      UInt64,
    price       Float64,
    day         Date
) ENGINE = SummingMergeTree(day,
           (day,
            banner_id,
            platform_id),
           8192);

CREATE TABLE kafka_clicks
(
    banner_id   UInt64,
    platform_id UInt64,
    view_id     String,
    price       Float64,
    created_at  UInt64
) ENGINE = Kafka('kafka-1:9092,kafka-2:9092,kafka-3:9092',
           'adclick.action.click',
           'ch-stat-clicks',
           'JSONEachRow');

CREATE MATERIALIZED VIEW consumer_clicks TO clicks AS
SELECT banner_id,
       platform_id,
       price,
       view_id,
       created_at,
       toDate(
               toDateTime(created_at)) AS day,
       toDateTime(
               created_at)             AS dt
FROM kafka_clicks;

CREATE MATERIALIZED VIEW clicks_daily_view TO clicks_daily AS
SELECT banner_id,
       platform_id,
       count(
           )          AS clicks,
       sum(
               price) AS price,
       day
FROM clicks
GROUP BY (
          day,
          banner_id,
          platform_id
             );