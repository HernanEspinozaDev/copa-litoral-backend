-- Migración inicial: Esquema completo de Copa Litoral
-- Versión: 001
-- Descripción: Crea todas las tablas principales del sistema

-- Extensión para generar UUIDs, útil para IDs si no quieres SERIAL
-- Descomenta la siguiente línea si necesitas UUIDs
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabla para los torneos (si hay múltiples ediciones)
CREATE TABLE IF NOT EXISTS torneos (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    anio INTEGER NOT NULL,
    fecha_inicio DATE,
    fecha_fin DATE,
    foto_url TEXT,
    frase_destacada TEXT,
    activo BOOLEAN DEFAULT TRUE, -- Para marcar el torneo actual
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla para las categorías
CREATE TABLE IF NOT EXISTS categorias (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla para los jugadores
CREATE TABLE IF NOT EXISTS jugadores (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    apellido VARCHAR(255) NOT NULL,
    telefono_wsp VARCHAR(50),
    contacto_visible_en_web BOOLEAN DEFAULT FALSE, -- Para controlar la visibilidad del contacto
    categoria_id INTEGER REFERENCES categorias(id) ON DELETE SET NULL, -- Si se borra una categoría, el jugador queda sin categoría
    club VARCHAR(255),
    estado_participacion VARCHAR(50) DEFAULT 'Activo', -- 'Activo', 'Eliminado', 'Inactivo'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla para los usuarios (administradores, jugadores que se loguean)
-- Aquí se almacenarán las credenciales para el login
CREATE TABLE IF NOT EXISTS usuarios (
    id SERIAL PRIMARY KEY,
    nombre_usuario VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE, -- Opcional, si quieres usar email para login/recuperación
    password_hash TEXT NOT NULL, -- Almacena el hash de la contraseña
    rol VARCHAR(50) NOT NULL DEFAULT 'jugador', -- 'jugador', 'administrador'
    jugador_id INTEGER UNIQUE REFERENCES jugadores(id) ON DELETE SET NULL, -- Opcional: vincular a un jugador existente
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tipo ENUM para el estado de los partidos
CREATE TYPE estado_partido_enum AS ENUM ('Pendiente', 'Agendado', 'Jugado', 'Walkover', 'Cancelado');

-- Tabla para los partidos
CREATE TABLE IF NOT EXISTS partidos (
    id SERIAL PRIMARY KEY,
    torneo_id INTEGER REFERENCES torneos(id) ON DELETE CASCADE, -- Si se borra un torneo, se borran sus partidos
    categoria_id INTEGER REFERENCES categorias(id) ON DELETE CASCADE,
    jugador1_id INTEGER REFERENCES jugadores(id) ON DELETE CASCADE NOT NULL,
    jugador2_id INTEGER REFERENCES jugadores(id) ON DELETE CASCADE NOT NULL,
    fase VARCHAR(100) NOT NULL, -- Ej: 'Fase de Grupos', 'Cuartos de Final', 'Semifinal', 'Final'
    fecha_agendada DATE,
    hora_agendada TIME,
    propuesta_fecha_j1 DATE, -- Propuesta de fecha por jugador 1
    propuesta_hora_j1 TIME, -- Propuesta de hora por jugador 1
    propuesta_fecha_j2 DATE, -- Propuesta de fecha por jugador 2
    propuesta_hora_j2 TIME, -- Propuesta de hora por jugador 2
    propuesta_aceptada_j1 BOOLEAN DEFAULT FALSE,
    propuesta_aceptada_j2 BOOLEAN DEFAULT FALSE,
    estado estado_partido_enum DEFAULT 'Pendiente',
    resultado_sets_j1 INTEGER, -- Número total de sets ganados por jugador 1
    resultado_sets_j2 INTEGER, -- Número total de sets ganados por jugador 2
    ganador_id INTEGER REFERENCES jugadores(id) ON DELETE SET NULL,
    perdedor_id INTEGER REFERENCES jugadores(id) ON DELETE SET NULL,
    resultado_aprobado BOOLEAN DEFAULT FALSE, -- Para aprobación de administrador
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla para almacenar los scores de cada set de un partido
CREATE TABLE IF NOT EXISTS sets_partido (
    id SERIAL PRIMARY KEY,
    partido_id INTEGER REFERENCES partidos(id) ON DELETE CASCADE,
    numero_set INTEGER NOT NULL,
    score_jugador1 INTEGER NOT NULL,
    score_jugador2 INTEGER NOT NULL,
    -- Opcional: para Tie-Breaks si son separados del score del set principal
    tie_break_j1 INTEGER,
    tie_break_j2 INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (partido_id, numero_set) -- Un partido no puede tener dos sets con el mismo número
);

-- Tabla para auspiciadores
CREATE TABLE IF NOT EXISTS auspiciadores (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    logo_url TEXT,
    enlace_web TEXT,
    descripcion TEXT,
    activo BOOLEAN DEFAULT TRUE,
    orden INTEGER DEFAULT 0, -- Para ordenar la visualización
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla para noticias/blog (opcional)
CREATE TABLE IF NOT EXISTS noticias (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL, -- Para URLs amigables
    contenido TEXT NOT NULL,
    fecha_publicacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    autor_id INTEGER REFERENCES usuarios(id) ON DELETE SET NULL,
    imagen_destacada_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla para campeones/historial
CREATE TABLE IF NOT EXISTS campeones (
    id SERIAL PRIMARY KEY,
    torneo_id INTEGER REFERENCES torneos(id) ON DELETE CASCADE,
    categoria_id INTEGER REFERENCES categorias(id) ON DELETE CASCADE,
    jugador_id INTEGER REFERENCES jugadores(id) ON DELETE SET NULL,
    anio INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (torneo_id, categoria_id, anio) -- Un campeón por torneo, categoría y año
);

-- Tabla para fotos destacadas o de partidos/finales
CREATE TABLE IF NOT EXISTS fotos_destacadas (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(255),
    descripcion TEXT,
    url TEXT NOT NULL,
    partido_id INTEGER REFERENCES partidos(id) ON DELETE SET NULL, -- Si la foto es de un partido específico
    torneo_id INTEGER REFERENCES torneos(id) ON DELETE SET NULL, -- Si la foto es de un torneo en general
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices para mejorar el rendimiento de las consultas
CREATE INDEX IF NOT EXISTS idx_partidos_torneo_categoria ON partidos (torneo_id, categoria_id);
CREATE INDEX IF NOT EXISTS idx_partidos_jugador1 ON partidos (jugador1_id);
CREATE INDEX IF NOT EXISTS idx_partidos_jugador2 ON partidos (jugador2_id);
CREATE INDEX IF NOT EXISTS idx_jugadores_categoria ON jugadores (categoria_id);
CREATE INDEX IF NOT EXISTS idx_usuarios_jugador ON usuarios (jugador_id);
CREATE INDEX IF NOT EXISTS idx_sets_partido ON sets_partido (partido_id);
CREATE INDEX IF NOT EXISTS idx_noticias_slug ON noticias (slug);
CREATE INDEX IF NOT EXISTS idx_campeones_torneo_categoria ON campeones (torneo_id, categoria_id);

-- Trigger para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplica el trigger a las tablas que necesiten 'updated_at'
CREATE TRIGGER set_timestamp_torneos BEFORE UPDATE ON torneos FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_categorias BEFORE UPDATE ON categorias FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_jugadores BEFORE UPDATE ON jugadores FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_usuarios BEFORE UPDATE ON usuarios FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_partidos BEFORE UPDATE ON partidos FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_sets_partido BEFORE UPDATE ON sets_partido FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_auspiciadores BEFORE UPDATE ON auspiciadores FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_noticias BEFORE UPDATE ON noticias FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_campeones BEFORE UPDATE ON campeones FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER set_timestamp_fotos_destacadas BEFORE UPDATE ON fotos_destacadas FOR EACH ROW EXECUTE FUNCTION update_timestamp();
