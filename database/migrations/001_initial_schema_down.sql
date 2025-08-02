-- Rollback de la migración inicial
-- Versión: 001
-- Descripción: Elimina todas las tablas del esquema inicial

-- Eliminar triggers
DROP TRIGGER IF EXISTS set_timestamp_fotos_destacadas ON fotos_destacadas;
DROP TRIGGER IF EXISTS set_timestamp_campeones ON campeones;
DROP TRIGGER IF EXISTS set_timestamp_noticias ON noticias;
DROP TRIGGER IF EXISTS set_timestamp_auspiciadores ON auspiciadores;
DROP TRIGGER IF EXISTS set_timestamp_sets_partido ON sets_partido;
DROP TRIGGER IF EXISTS set_timestamp_partidos ON partidos;
DROP TRIGGER IF EXISTS set_timestamp_usuarios ON usuarios;
DROP TRIGGER IF EXISTS set_timestamp_jugadores ON jugadores;
DROP TRIGGER IF EXISTS set_timestamp_categorias ON categorias;
DROP TRIGGER IF EXISTS set_timestamp_torneos ON torneos;

-- Eliminar función de trigger
DROP FUNCTION IF EXISTS update_timestamp();

-- Eliminar tablas en orden inverso (respetando foreign keys)
DROP TABLE IF EXISTS fotos_destacadas;
DROP TABLE IF EXISTS campeones;
DROP TABLE IF EXISTS noticias;
DROP TABLE IF EXISTS auspiciadores;
DROP TABLE IF EXISTS sets_partido;
DROP TABLE IF EXISTS partidos;
DROP TABLE IF EXISTS usuarios;
DROP TABLE IF EXISTS jugadores;
DROP TABLE IF EXISTS categorias;
DROP TABLE IF EXISTS torneos;

-- Eliminar tipos ENUM
DROP TYPE IF EXISTS estado_partido_enum;
