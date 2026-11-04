# Prompt para Windsurf: Landing Page + Autenticación (Split Layout)

Actúa como un desarrollador Frontend experto en **React** y **Tailwind CSS**. Necesito refactorizar la pantalla de inicio de la aplicación para convertirla en una **Landing Page de presentación dividida en dos columnas (Split Layout)** responsiva.

---

## 1. Columna Izquierda: Información del Concurso (Landing Page)

Diseña una sección informativa, moderna y limpia.

### Header / Título Principal

- **Título:** `Concurso Público de Méritos y Oposición BIESS`
- **Subtítulo:** `Selección de Miembros Principales y Alternos del Directorio del BIESS (Resolución No. C.D. 701)`

### Bloque 1: Requisitos Clave

Usar iconos o tarjetas pequeñas para representar cada requisito.

1. **Nacionalidad:** Poseer nacionalidad ecuatoriana.
2. **Derechos Políticos:** Estar en pleno goce de derechos de participación política y sin inhabilidades legales.
3. **Formación y Experiencia:** Título de Tercer o Cuarto Nivel registrado en SENESCYT y trayectoria acreditada en Finanzas, Mercado de Capitales, Administración de Empresas, Economía, Derecho o áreas afines.
4. **Condición del Sector:** Acreditar la calidad activa de afiliado, jubilado o empleador del IESS según el sector de postulación.

### Bloque 2: Pasos del Proceso

Diseñar como una línea de tiempo vertical o lista de pasos numerados del 1 al 4.

1. **Registro:** Valida tu identidad ingresando tu cédula y fecha de expedición o código dactilar.
2. **Formulario Digital:** Completa tus datos personales, experiencia laboral y adjunta los respaldos en PDF.
3. **Envío Digital:** Acepta la Declaración Jurada, genera tu Código Único de Postulación y envía el formulario.
4. **Entrega Física:** Imprime el formulario completado, coloca tu firma manuscrita y entrégalo físicamente en las dependencias del IESS.

---

## 2. Columna Derecha: Panel de Autenticación

Diseña una tarjeta (`Card`) flotante o destacada a la derecha, con una sombra limpia, que contenga los accesos al sistema.

### Título de la Tarjeta

`Acceso a la Plataforma`

### Pestañas (Tabs) o Botones

Implementar dos opciones para alternar entre los formularios:

#### Iniciar Sesión

Formulario con los siguientes campos:

- Cédula.
- Contraseña.

#### Registrarse

Flujo inicial de registro con:

- Cédula.
- Selector de Fecha de Expedición / Código Dactilar.

### Enlace al pie

`¿Olvidaste tu contraseña?`

---

## 3. Requisitos Técnicos y de Estilo

### Layout Responsivo

- **Escritorio (`lg:grid-cols-12` o `flex-row`):** Distribución en dos columnas:
  - 7 columnas para la información del concurso (izquierda).
  - 5 columnas para la tarjeta de autenticación (derecha).
- **Dispositivos móviles (`sm` / `md`):** Distribución apilada verticalmente en una sola columna.

### Estilos

- Utilizar **Tailwind CSS** con tonos institucionales:
  - Azules.
  - Grises oscuros.
  - Blancos.
  - Bordes suaves (`rounded-xl`).
- Mantener una apariencia institucional, moderna, limpia y profesional.

### Modularidad

- Mantener componentes limpios y modulares.
- Reutilizar la lógica o el estado de autenticación existente.
- Evitar duplicar lógica de autenticación.
- Mantener la separación de responsabilidades entre presentación y lógica de negocio.