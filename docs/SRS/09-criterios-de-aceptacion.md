# 09. Criterios de aceptación — SRS

## 9.1 Convenciones

Los criterios se escriben en formato Gherkin para que puedan convertirse en pruebas automatizadas. Cada escenario referencia los requisitos y reglas que verifica.

| Elemento | Significado |
|---|---|
| `Característica` | Capacidad de negocio. |
| `Antecedentes` | Estado común a los escenarios. |
| `Escenario` | Caso concreto y verificable. |
| `Dado` | Precondición. |
| `Cuando` | Acción del actor o del sistema. |
| `Entonces` | Resultado observable. |

---

## 9.2 Registro exitoso

**Característica:** Registro de postulante con identidad validada.  
**Referencias:** RF-EF11-02, RF-EF11-03; RN-02, RN-03, RN-04, RN-05.

```gherkin
Escenario: Registro exitoso con fecha de expedición correcta
  Dado que el postulante accede a la opción "Registrarse"
  Y la fase de postulación está habilitada
  Cuando ingresa una cédula válida y su fecha de expedición correcta
  Entonces el sistema valida la identidad con Registro Civil
  Y muestra los datos verificados en modo lectura
  Cuando ingresa un correo único y una contraseña que cumple la política
  Entonces el sistema crea la cuenta con estado ACTIVO
  Y muestra el mensaje "Registro exitoso"
  Y redirige a la pantalla de inicio de sesión
```

**Criterios adicionales:**

- La cédula no puede registrarse dos veces.
- El correo no puede registrarse dos veces.
- La contraseña no se almacena en claro.
- El registro exitoso no inicia sesión automáticamente.

---

## 9.3 Identidad inválida y bloqueo de registro

**Característica:** Control de intentos fallidos de registro.  
**Referencias:** RF-EF11-02, RF-EF11-04; RN-06, RN-08.

```gherkin
Escenario: Bloqueo tras tres intentos fallidos de validación
  Dado que el postulante está en la pantalla de registro
  Cuando ingresa un dato de validación incorrecto por primera vez
  Entonces el sistema muestra "La información ingresada es incorrecta"
  Y registra el intento fallido
  Cuando repite el error por segunda y tercera vez
  Entonces el sistema muestra "Registro fallido, inténtelo nuevamente en 10 minutos"
  Y bloquea el registro durante 10 minutos
  Y no permite nuevos intentos hasta que expire el bloqueo
```

```gherkin
Escenario: El bloqueo expira y permite un nuevo intento
  Dado que el postulante fue bloqueado por 3 intentos fallidos
  Cuando transcurren 10 minutos
  Entonces el sistema permite un nuevo intento de validación
  Y el contador de intentos se reinicia
```

```gherkin
Escenario: Un error técnico no consume intento
  Dado que el servicio de Registro Civil no responde
  Cuando el postulante intenta validar su identidad
  Entonces el sistema muestra "No es posible validar su identidad en este momento"
  Y no incrementa el contador de intentos fallidos
```

---

## 9.4 Inicio de sesión y bloqueo

**Característica:** Autenticación segura.  
**Referencias:** RF-EF11-05, RF-EF11-06; RN-05, RN-07.

```gherkin
Escenario: Inicio de sesión exitoso
  Dado que el postulante tiene una cuenta activa
  Cuando ingresa su cédula y contraseña correctas
  Entonces el sistema emite un token de acceso y un refresh token
  Y actualiza la fecha de último acceso
  Y redirige al formulario o al resumen de su postulación
```

```gherkin
Escenario: Credenciales incorrectas sin revelar información
  Dado que el postulante ingresa una contraseña incorrecta
  Cuando el sistema valida las credenciales
  Entonces muestra "Sus credenciales de ingreso son incorrectas"
  Y no indica si la cédula existe o no
```

```gherkin
Escenario: Bloqueo tras tres intentos fallidos de login
  Dado que el postulante falla el login tres veces consecutivas
  Cuando intenta un cuarto ingreso
  Entonces el sistema muestra "Cuenta bloqueada temporalmente. Intente nuevamente en 10 minutos"
  Y responde con HTTP 423
  Y no permite el acceso hasta que expire el bloqueo
```

---

## 9.5 Recuperación de contraseña

**Característica:** Recuperación segura de credenciales.  
**Referencias:** RF-EF11-07, RF-EF11-08; RN-02, RN-09.

```gherkin
Escenario: Solicitud de recuperación con identidad válida
  Dado que el postulante olvidó su contraseña
  Cuando ingresa su cédula y un dato de validación correcto
  Entonces el sistema genera un token de un solo uso
  Y envía un correo con el enlace de restablecimiento
  Y muestra un mensaje genérico de confirmación
```

```gherkin
Escenario: Restablecimiento de contraseña
  Dado que el postulante recibió un enlace válido
  Cuando ingresa una nueva contraseña que cumple la política
  Entonces el sistema actualiza la contraseña
  Y invalida el token y las sesiones anteriores
  Y redirige al login con un mensaje de éxito
```

```gherkin
Escenario: Token expirado
  Dado que el enlace de recuperación tiene más de 30 minutos
  Cuando el postulante intenta usarlo
  Entonces el sistema muestra "El enlace de recuperación expiró. Solicite uno nuevo"
  Y no permite cambiar la contraseña
```

---

## 9.6 Primer ingreso y redirección al formulario

**Característica:** Continuidad del proceso de postulación.  
**Referencias:** RF-EF11-10, RF-EF12-01; RN-11, RN-12.

```gherkin
Escenario: Primer ingreso crea el borrador
  Dado que el postulante inicia sesión por primera vez
  Y no tiene una postulación vigente
  Cuando el sistema consulta su postulación actual
  Entonces crea un borrador en estado BORRADOR
  Y lo redirige a la primera sección del formulario
```

```gherkin
Escenario: Reanudación de un borrador
  Dado que el postulante tiene un borrador con la sección de experiencia guardada
  Cuando vuelve a iniciar sesión
  Entonces el sistema lo redirige a la última sección guardada
  Y conserva todos los datos previamente ingresados
```

---

## 9.7 Restricción de nacionalidad

**Característica:** Elegibilidad inicial por nacionalidad.  
**Referencias:** RF-EF12-04; RN-01.

```gherkin
Escenario: Postulante no ecuatoriano no puede continuar
  Dado que Registro Civil devuelve una nacionalidad distinta de Ecuador
  Cuando el postulante intenta guardar la información personal
  Entonces el sistema muestra "Su postulación no puede continuar porque la nacionalidad registrada no es ecuatoriana"
  Y bloquea la sección
  Y no permite guardar ni enviar
  Y registra el evento en auditoría
```

```gherkin
Escenario: Postulante ecuatoriano continúa
  Dado que Registro Civil devuelve nacionalidad Ecuador
  Cuando el postulante guarda la información personal
  Entonces el sistema permite continuar a la sección de domicilio
```

---

## 9.8 Cascada geográfica

**Característica:** Selección de ubicación.  
**Referencias:** RF-EF12-06, RF-EF12-07, RF-CAT-01 a RF-CAT-04; RN-16, RN-39.

```gherkin
Escenario: Cascada para Ecuador
  Dado que el postulante selecciona Ecuador como país
  Cuando el sistema carga las provincias
  Entonces muestra las 24 provincias de Ecuador
  Cuando selecciona una provincia
  Entonces carga únicamente los cantones de esa provincia
```

```gherkin
Escenario: Ubicación en el extranjero
  Dado que el postulante selecciona un país distinto de Ecuador
  Cuando el sistema muestra el formulario
  Entonces habilita los campos de provincia/estado y ciudad para ingreso manual
  Y no exige cantón del catálogo ecuatoriano
```

```gherkin
Escenario: Cambio de país limpia dependencias
  Dado que el postulante seleccionó Ecuador y una provincia
  Cuando cambia el país a otro distinto
  Entonces el sistema limpia la provincia y el cantón seleccionados
  Y habilita la captura manual
```

---

## 9.9 Guardado de progreso

**Característica:** Persistencia por secciones.  
**Referencias:** RF-EF12-02; RN-14, RN-36.

```gherkin
Escenario: Guardar una sección incompleta
  Dado que el postulante completó parcialmente la sección de experiencia
  Cuando pulsa "Guardar progreso"
  Entonces el sistema guarda los campos válidos de la sección
  Y muestra "Progreso guardado"
  Y actualiza el porcentaje de avance
  Y no exige completar las demás secciones
```

```gherkin
Escenario: Conflicto de versión
  Dado que la postulación fue modificada en otra pestaña
  Cuando el postulante guarda una sección con una versión desactualizada
  Entonces el sistema responde HTTP 409
  Y muestra "La postulación fue modificada en otra sesión. Recargue la página"
```

---

## 9.10 Elegibilidad por sector

**Característica:** Validación de calidad laboral.  
**Referencias:** RF-EF12-09, RF-EF12-10; RN-19, RN-20.

```gherkin
Escenario: Postulante afiliado elegible
  Dado que el postulante selecciona el cargo de Afiliados
  Y Historial Laboral confirma que es afiliado
  Cuando confirma el cargo
  Entonces el sistema muestra "Postulante afiliado al IESS"
  Y habilita la pregunta de respaldo organizacional
```

```gherkin
Escenario: Postulante no afiliado
  Dado que el postulante selecciona el cargo de Afiliados
  Y Historial Laboral indica que no es afiliado
  Cuando confirma el cargo
  Entonces el sistema muestra el mensaje institucional de cierre
  Y marca la postulación como NO_ELEGIBLE
  Y no permite continuar
```

```gherkin
Escenario: Servicio de Historial Laboral no disponible
  Dado que el servicio de Historial Laboral no responde
  Cuando el postulante confirma el cargo
  Entonces el sistema muestra "No es posible validar su calidad laboral en este momento"
  Y no marca la postulación como NO_ELEGIBLE
  Y permite reintentar
```

---

## 9.11 Validación de documentos

**Característica:** Carga segura de respaldos.  
**Referencias:** RF-EF12-17; RN-27, RN-28.

```gherkin
Escenario: PDF válido dentro del límite
  Dado que el postulante selecciona un archivo PDF de 1 MB
  Cuando lo carga en la sección de formación
  Entonces el sistema valida el MIME real y el tamaño
  Y almacena el archivo con un nombre UUID
  Y muestra el documento en la lista
```

```gherkin
Escenario: Archivo que no es PDF
  Dado que el postulante selecciona un archivo .docx renombrado a .pdf
  Cuando intenta cargarlo
  Entonces el sistema rechaza la carga
  Y muestra "Solo se permiten archivos PDF"
```

```gherkin
Escenario: Archivo que supera el límite
  Dado que el postulante selecciona un PDF de 5 MB para formación
  Cuando intenta cargarlo
  Entonces el sistema rechaza la carga
  Y muestra "El archivo supera el tamaño máximo permitido"
```

```gherkin
Escenario: Límite total de 50 MB
  Dado que el postulante ya cargó 49 MB en documentos
  Cuando intenta cargar un PDF de 2 MB
  Entonces el sistema rechaza la carga
  Y muestra "Ha superado el límite total de 50 MB"
```

---

## 9.12 Guardar postulación incompleta

**Característica:** Validación integral antes de guardar.  
**Referencias:** RF-EF12-18; RN-29, RN-30.

```gherkin
Escenario: Postulación con secciones pendientes
  Dado que el postulante no completó la sección de inhabilidades
  Cuando pulsa "Guardar postulación"
  Entonces el sistema responde HTTP 422
  Y muestra la lista de secciones pendientes
  Y no cambia el estado a LISTA_PARA_ENVIO
```

```gherkin
Escenario: Consentimiento no aceptado
  Dado que el postulante completó todas las secciones
  Cuando intenta guardar sin aceptar los cinco consentimientos
  Entonces el sistema no permite guardar
  Y resalta los consentimientos pendientes
```

---

## 9.13 Envío exitoso

**Característica:** Envío atómico de la postulación.  
**Referencias:** RF-EF12-19, RF-EF12-20, RF-EF12-23; RN-30 a RN-32, RN-37.

```gherkin
Escenario: Envío de una postulación completa
  Dado que la postulación está en estado LISTA_PARA_ENVIO
  Y el postulante aceptó los cinco consentimientos
  Cuando confirma "Enviar postulación"
  Entonces el sistema genera un código único con formato BIESS-2026-{SECTOR}-{SECUENCIA}
  Y genera el PDF con todas las secciones
  Y cambia el estado a ENVIADA
  Y encola el correo con el PDF adjunto
  Y muestra el código y el enlace de descarga
```

```gherkin
Escenario: Envío idempotente
  Dado que la postulación ya fue enviada con el código BIESS-2026-AFI-000123
  Cuando el postulante intenta enviarla nuevamente
  Entonces el sistema responde HTTP 200
  Y devuelve el mismo código BIESS-2026-AFI-000123
  Y no duplica la postulación ni el correo
```

```gherkin
Escenario: Falla la generación del PDF
  Dado que el generador de PDF presenta un error
  Cuando el postulante confirma el envío
  Entonces el sistema no cambia el estado a ENVIADA
  Y muestra "No fue posible generar su formulario. Intente nuevamente"
  Y no genera un código único
```

---

## 9.14 Inmutabilidad después del envío

**Característica:** Protección de la postulación enviada.  
**Referencias:** RF-EF12-22; RN-31.

```gherkin
Escenario: Edición después del envío
  Dado que la postulación está en estado ENVIADA
  Cuando el postulante intenta modificar una experiencia
  Entonces el sistema responde HTTP 409
  Y muestra "La postulación ya fue enviada y no puede modificarse"
  Y la interfaz muestra el resumen en modo lectura
```

---

## 9.15 Resiliencia de Redis

**Característica:** Degradación controlada.  
**Referencias:** RNF-DIS-04; RF-CAT-04.

```gherkin
Escenario: Catálogos con Redis caído
  Dado que Redis no está disponible
  Cuando el postulante consulta las provincias
  Entonces el sistema consulta PostgreSQL
  Y devuelve la lista correctamente
  Y registra una advertencia en los logs
```

```gherkin
Escenario: Bloqueo de autenticación con Redis caído
  Dado que Redis no está disponible
  Cuando el postulante intenta validar su identidad
  Entonces el sistema no permite intentos ilimitados
  Y muestra un mensaje de indisponibilidad temporal
  Y registra el incidente
```

---

## 9.16 Rollback transaccional

**Característica:** Integridad de la postulación.  
**Referencias:** RNF-DAT-05; RF-EF12-19.

```gherkin
Escenario: Falla la inserción del documento PDF
  Dado que la transacción de envío actualizó el estado
  Cuando falla la inserción del registro del PDF
  Entonces el sistema ejecuta ROLLBACK
  Y la postulación permanece en LISTA_PARA_ENVIO
  Y no existe un código único generado
  Y no se encola ningún correo
```

---

## 9.17 Accesibilidad y feedback

**Característica:** Usabilidad de los flujos críticos.  
**Referencias:** RNF-USA-01 a RNF-USA-05.

```gherkin
Escenario: Navegación por teclado en el formulario
  Dado que el postulante usa solo el teclado
  Cuando recorre el formulario con Tab
  Entonces el foco es visible en cada campo
  Y los mensajes de error se anuncian
  Y puede guardar la sección sin usar el mouse
```

```gherkin
Escenario: Cambios sin guardar
  Dado que el postulante modificó un campo y no guardó
  Cuando intenta salir de la sección
  Entonces el sistema muestra una advertencia
  Y permite quedarse o descartar los cambios
```

---

## 9.18 Validación algorítmica de cédula ecuatoriana

**Característica:** Validación sintáctica de cédula (Módulo 10).  
**Referencias:** RF-VAL-01, RN-01, RN-41, ADR-013.

```gherkin
Escenario: Cédula con formato y dígito verificador válido
  Dado que el postulante ingresa la cédula "1710034065"
  Cuando el sistema ejecuta la validación algorítmica
  Entonces comprueba que tiene 10 dígitos numéricos
  Y verifica que el código provincial "17" es válido
  Y calcula el dígito verificador aplicando Módulo 10 obteniendo "5"
  Y declara la cédula como sintácticamente válida
```

```gherkin
Escenario: Cédula con dígito verificador fraudulento o erróneo
  Dado que el postulante ingresa la cédula "1710034069"
  Cuando el sistema ejecuta la validación algorítmica
  Entonces detecta que el residuo esperado no coincide con el dígito verificador
  Y rechaza la petición con código HTTP 422
  Y muestra el mensaje "El número de cédula ingresado no es válido"
  Y no realiza ninguna petición al Registro Civil
```

---

## 9.19 Procesamiento de pago con tarjeta y emisión de factura

> **Retirado (03/11/2026).** La postulación al Directorio del BIESS es gratuita: no se cobra arancel ni tasa y no se emiten facturas. Esta sección se conserva solo como registro. Ver `db/migrations/002_eliminar_pagos.sql` y SRP-010 en `docs/TICKETS.md`.

**Característica:** Pago con tarjeta y facturación electrónica (Mockeable e Implementable).  
**Referencias:** RF-PAG-01, RF-PAG-02, RN-42, RN-43, ADR-014.

```gherkin
Escenario: Pago exitoso con tarjeta y emisión de factura
  Dado que el postulante se encuentra en el paso de pago de arancel
  Cuando ingresa una tarjeta de prueba válida con CVV y expiración futura
  Y confirma el pago por el monto de 25.00 USD
  Entonces el validador de Luhn aprueba el número de tarjeta
  Y el adaptador de pago (Mock/Gateway) autoriza la transacción generando código de autorización
  Y se registra la transacción de pago en estado APROBADO de forma atómica en PostgreSQL
  Y el servicio de facturación genera la factura con clave de acceso de 49 dígitos del SRI
  Y se encola en el Outbox el envío del correo electrónico con el PDF de la factura
  Y el postulante visualiza la confirmación con el número de factura y código de autorización
```

```gherkin
Escenario: Pago declinado por tarjeta rechazada o fondos insuficientes
  Dado que el postulante ingresa una tarjeta de prueba configurada para simular fondos insuficientes
  Cuando confirma el pago
  Entonces el adaptador de pago responde declinado
  Y el sistema registra el intento fallido en base de datos
  Y responde HTTP 402 con mensaje "La transacción fue declinada por su entidad financiera"
  Y no se emite ninguna factura ni se encola correo alguno
```
