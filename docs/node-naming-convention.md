# Convención de Nomenclatura de Nodos SECOP

## Formato General
```
[secop]-[entidad]-[tipo]-[ubicacion]
```

## Ejemplos por Tipo de Entidad

### Entidades de Gobierno Nacional
```bash
# DNP (Departamento Nacional de Planeación)
NODE_ID=secop-government-central-bogota
ENTITY_TYPE=GOVERNMENT

# Ministerios y entidades nacionales
NODE_ID=secop-government-minhacienda-bogota
ENTITY_TYPE=GOVERNMENT
```

### DNP (Departamento Nacional de Planeación) - Legacy
```bash
NODE_ID=secop-dnp-central-bogota
ENTITY_TYPE=DNP
```

### Municipios
```bash
# Alcaldía de Medellín
NODE_ID=secop-medellin-alcaldia-main
ENTITY_TYPE=MUNICIPALITY

# Alcaldía de Bogotá
NODE_ID=secop-bogota-alcaldia-main
ENTITY_TYPE=MUNICIPALITY

# Alcaldía de Cali
NODE_ID=secop-cali-alcaldia-main
ENTITY_TYPE=MUNICIPALITY
```

### Departamentos
```bash
# Gobernación de Antioquia
NODE_ID=secop-antioquia-gobernacion-main
ENTITY_TYPE=DEPARTMENT

# Gobernación del Valle
NODE_ID=secop-valle-gobernacion-main
ENTITY_TYPE=DEPARTMENT

# Gobernación de Cundinamarca
NODE_ID=secop-cundinamarca-gobernacion-main
ENTITY_TYPE=DEPARTMENT
```

### Ministerios
```bash
# Ministerio de Hacienda
NODE_ID=secop-minhacienda-central-bogota
ENTITY_TYPE=MINISTRY

# Ministerio de Educación
NODE_ID=secop-mineducacion-central-bogota
ENTITY_TYPE=MINISTRY
```

### Entidades de Control
```bash
# Contraloría General de la República
NODE_ID=secop-contraloria-general-bogota
ENTITY_TYPE=CONTROL

# Procuraduría General de la Nación
NODE_ID=secop-procuraduria-general-bogota
ENTITY_TYPE=CONTROL

# Contraloría de Medellín
NODE_ID=secop-contraloria-medellin-main
ENTITY_TYPE=CONTROL
```

## Migración desde V1

### V1 (Anterior)
```yaml
NODE_ID=DNP
NODE_ID=MEDELLIN
NODE_ID=BOGOTA
NODE_ID=CONTROL
```

### V2 (Actual)
```bash
NODE_ID=secop-dnp-central-bogota
NODE_ID=secop-medellin-alcaldia-main
NODE_ID=secop-bogota-alcaldia-main
NODE_ID=secop-contraloria-general-bogota
```

## Beneficios del Nuevo Sistema

1. **Identificación SECOP**: Prefijo claro que identifica el sistema
2. **Claridad**: Identifica inmediatamente la entidad, tipo y ubicación
3. **Escalabilidad**: Soporta múltiples nodos por entidad
4. **Organización**: Facilita la administración de la red
5. **Compatibilidad**: Funciona con el sistema de descubrimiento dinámico
