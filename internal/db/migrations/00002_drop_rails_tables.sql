-- +goose Up
DROP TABLE IF EXISTS
  solid_queue_blocked_executions,
  solid_queue_claimed_executions,
  solid_queue_failed_executions,
  solid_queue_ready_executions,
  solid_queue_recurring_executions,
  solid_queue_scheduled_executions,
  solid_queue_jobs,
  solid_queue_pauses,
  solid_queue_processes,
  solid_queue_recurring_tasks,
  solid_queue_semaphores,
  schema_migrations,
  ar_internal_metadata;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
  RAISE EXCEPTION 'Rails table removal is irreversible; restore a backup to roll back';
END $$;
-- +goose StatementEnd