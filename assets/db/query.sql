-- name: TenderCount :one
SELECT count(*) FROM tender
WHERE deleted = 0
    LIMIT 1;

-- name: TenderListAll :many
SELECT uid, uuid, created_at, updated_at, source, link, title, unix_date, categories, description, guid, json_version, json, attempt, error, publish_at, closing_at, deleted
FROM tender
WHERE deleted = 0
ORDER BY closing_at DESC;

-- name: TenderListSearchable :many
SELECT uid, uuid, created_at, updated_at, source, link, title, unix_date, categories, description, guid, json_version, json, attempt, error, publish_at, closing_at, deleted
FROM tender
WHERE deleted = 0
  AND closing_at > ?
ORDER BY closing_at ASC;

-- name: TenderCreate :one
INSERT INTO tender (uuid, created_at, updated_at, source, link, title, unix_date, categories, description, guid, json_version, json, attempt, error, publish_at, closing_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
    returning *;

-- name: TenderGetByLink :one
SELECT uid, uuid, created_at, updated_at, source, link, title, unix_date, categories, description, guid, json_version, json, attempt, error, publish_at, closing_at, deleted
FROM tender
WHERE link = ?;

-- name: TenderGetByUuid :one
SELECT uid, uuid, created_at, updated_at, source, link, title, unix_date, categories, description, guid, json_version, json, attempt, error, publish_at, closing_at, deleted
FROM tender
WHERE uuid = ?;

-- name: TenderExistsByLink :one
SELECT COUNT(*)
FROM tender
WHERE link = ?;

-- name: TenderUpdateJson :exec
UPDATE tender SET
                  json_version = ?,
                  json = ?
WHERE uid = ?;
