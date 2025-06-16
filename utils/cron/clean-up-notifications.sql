DELETE FROM notifications
WHERE deleted_after < (now() at time zone 'utc');
