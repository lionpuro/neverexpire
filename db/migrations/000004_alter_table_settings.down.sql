alter table settings
drop constraint ck_settings_webhook_null_link;

alter table settings
drop column webhook_provider;
