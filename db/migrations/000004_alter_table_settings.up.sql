alter table settings
add webhook_provider text;

/* make sure webhook_url cannot exist without webhook_provider */
alter table settings
add constraint ck_settings_webhook_null_link
check (
	(webhook_provider is null and webhook_url = '')
	or (webhook_provider is not null and webhook_url != '')
);
