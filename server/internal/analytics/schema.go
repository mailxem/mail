package analytics

import "gorm.io/gorm"

// Install is transactional and idempotent. Database triggers cover bulk writes,
// UpdateColumn, and imports that intentionally skip GORM hooks.
func Install(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	return db.Exec(`
 CREATE TABLE IF NOT EXISTS analytics_coverage (key text PRIMARY KEY, started_at timestamptz NOT NULL);
 CREATE TABLE IF NOT EXISTS subscription_events (
   id bigserial PRIMARY KEY, team_id uuid NOT NULL, contact_id uuid NOT NULL,
   list_id uuid NOT NULL, email text NOT NULL, status text NOT NULL,
   is_deleted boolean NOT NULL, source text NOT NULL, kind text NOT NULL,
   occurred_at timestamptz NOT NULL DEFAULT clock_timestamp()
 );
 CREATE INDEX IF NOT EXISTS subscription_events_scope ON subscription_events(team_id, occurred_at, contact_id);
 CREATE INDEX IF NOT EXISTS subscription_events_contact ON subscription_events(team_id, contact_id, occurred_at DESC, id DESC);
 CREATE OR REPLACE FUNCTION xem_subscription_history() RETURNS trigger LANGUAGE plpgsql AS $$
 DECLARE r contacts; event_kind text;
 BEGIN
   IF TG_OP = 'UPDATE' AND ROW(OLD.team_id,OLD.list_id,OLD.email,OLD.status,OLD.is_deleted)
       IS NOT DISTINCT FROM ROW(NEW.team_id,NEW.list_id,NEW.email,NEW.status,NEW.is_deleted) THEN RETURN NEW; END IF;
   IF TG_OP = 'DELETE' THEN r := OLD; r.is_deleted := true; ELSE r := NEW; END IF;
   IF TG_OP = 'UPDATE' AND OLD.team_id IS DISTINCT FROM NEW.team_id THEN
     INSERT INTO subscription_events(team_id,contact_id,list_id,email,status,is_deleted,source,kind)
     VALUES(OLD.team_id,OLD.id,OLD.list_id,lower(trim(OLD.email)),OLD.status,true,'unknown','moved');
   END IF;
   r.is_deleted := COALESCE(r.is_deleted,false) OR COALESCE((SELECT is_deleted FROM mailing_lists WHERE id=r.list_id AND team_id=r.team_id),true);
   event_kind := lower(TG_OP);
   INSERT INTO subscription_events(team_id,contact_id,list_id,email,status,is_deleted,source,kind)
   VALUES(r.team_id,r.id,r.list_id,lower(trim(r.email)),r.status,COALESCE(r.is_deleted,false),
     CASE WHEN r.import_id IS NOT NULL THEN 'import' WHEN r.metadata->>'source' = 'form' THEN 'form' ELSE 'unknown' END,event_kind);
   IF TG_OP = 'DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
 END $$;
 DROP TRIGGER IF EXISTS xem_subscription_history ON contacts;
 CREATE TRIGGER xem_subscription_history AFTER INSERT OR UPDATE OR DELETE ON contacts
 FOR EACH ROW EXECUTE FUNCTION xem_subscription_history();
 CREATE OR REPLACE FUNCTION xem_list_subscription_history() RETURNS trigger LANGUAGE plpgsql AS $$
 BEGIN
   IF OLD.is_deleted IS NOT DISTINCT FROM NEW.is_deleted THEN RETURN NEW; END IF;
   INSERT INTO subscription_events(team_id,contact_id,list_id,email,status,is_deleted,source,kind)
   SELECT c.team_id,c.id,c.list_id,lower(trim(c.email)),c.status,COALESCE(c.is_deleted,false) OR NEW.is_deleted,
   CASE WHEN c.import_id IS NOT NULL THEN 'import' WHEN c.metadata->>'source'='form' THEN 'form' ELSE 'unknown' END,
   'list_status' FROM contacts c WHERE c.list_id=NEW.id AND c.team_id=NEW.team_id;
   RETURN NEW;
 END $$;
 DROP TRIGGER IF EXISTS xem_list_subscription_history ON mailing_lists;
 CREATE TRIGGER xem_list_subscription_history AFTER UPDATE ON mailing_lists FOR EACH ROW EXECUTE FUNCTION xem_list_subscription_history();
 WITH started AS (
   INSERT INTO analytics_coverage(key,started_at) VALUES('subscription-history',clock_timestamp())
   ON CONFLICT DO NOTHING RETURNING started_at
 )
 INSERT INTO subscription_events(team_id,contact_id,list_id,email,status,is_deleted,source,kind,occurred_at)
 SELECT c.team_id,c.id,c.list_id,lower(trim(c.email)),c.status,COALESCE(c.is_deleted,false) OR COALESCE(l.is_deleted,true),
   CASE WHEN c.import_id IS NOT NULL THEN 'import' WHEN c.metadata->>'source' = 'form' THEN 'form' ELSE 'unknown' END,
   'baseline',s.started_at FROM contacts c LEFT JOIN mailing_lists l ON l.id=c.list_id AND l.team_id=c.team_id CROSS JOIN started s;
 CREATE INDEX IF NOT EXISTS analytics_contacts_scope ON contacts(team_id,list_id,is_deleted);
 CREATE INDEX IF NOT EXISTS analytics_contacts_recipient ON contacts(team_id,lower(trim(email)));
 CREATE INDEX IF NOT EXISTS analytics_emails_scope ON emails(team_id,sent_at,campaign_id) WHERE test = false;
 CREATE INDEX IF NOT EXISTS analytics_tracking_message ON email_trackings(email_id,timestamp,event) WHERE is_deleted = false;
 `).Error
}
