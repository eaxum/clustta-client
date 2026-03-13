DROP TRIGGER IF EXISTS entity_path_update;

CREATE TRIGGER entity_path_update 
AFTER UPDATE OF name, parent_id ON entity
FOR EACH ROW
WHEN OLD.name != NEW.name OR OLD.parent_id != NEW.parent_id
BEGIN
    -- Recalculate this entity's path
    UPDATE entity
    SET entity_path =
        CASE
        WHEN NEW.parent_id IS NULL THEN '/' || NEW.name || '/'
        ELSE COALESCE(
            (SELECT entity_path || NEW.name || '/' FROM entity WHERE id = NEW.parent_id),
            '/' || NEW.name || '/'
        )
        END
    WHERE id = NEW.id;

    -- Recalculate all descendant paths
    UPDATE entity
    SET entity_path =
        (SELECT entity_path FROM entity WHERE id = NEW.id) || substr(entity_path, length(OLD.entity_path) + 1)
    WHERE entity_path LIKE OLD.entity_path || '%'
        AND id != NEW.id;
END;
