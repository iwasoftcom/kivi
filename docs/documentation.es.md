Kivi

la base de datos que muestra sus recibos · v1.1.0 · formato v2

Una base de datos de registro de eventos en Go puro — un único binario estático capaz de rendir cuentas de cada respuesta que da.

## Qué es Kivi

Kivi almacena solo **eventos** — registros inmutables, encadenados por hash. Todo lo que parece estado actual (una tabla, un grafo, una serie temporal, un índice vectorial) es una **vista**: un plegado sobre esos eventos, recompilado bajo demanda y descartable en cualquier momento. Cada respuesta lleva una **traza** (los números de registro que la establecen) y un **alcance** (hasta dónde miró en la historia), y `why` devuelve los propios registros-recibo. Un dato ausente es un rechazo honesto, nunca una invención.

## Los tres principios

**Solo los eventos son duraderos.**  
No hay UPDATE. Los hechos nuevos se anexan; el pasado nunca se reescribe. La historia es de primera clase.

**Las representaciones se compilan.**  
Borrar `derived/` siempre es legítimo: las respuestas no cambian, solo la siguiente consulta es más lenta.

**Ninguna respuesta sin traza.**  
Una respuesta sin traza es irrepresentable — en el núcleo y en cada SDK. Nada se inventa.

## En qué se diferencia de una base de datos normal

|  | Una base de datos típica | Kivi |
| --- | --- | --- |
| Qué es duradero | El estado actual; UPDATE destruye el pasado | Los eventos; el estado se recompila de ellos |
| Respuestas | Valores desnudos | Valor + traza + alcance; `why` obtiene los recibos |
| Punto en el tiempo | Instantáneas y maquinaria extra | Gratis por diseño: plegar hasta el registro N (`--as-of`) |
| Integridad | Supuesta | Verificada: cadena SHA-256 + sellos Ed25519; los clientes reverifican el servidor |
| Borrado | Fila fuera, historia fuera | Borrado cripto: clave destruida, bytes intactos, borrado registrado |

## Inicio rápido

Ejecute el servidor (binario estático único o imagen de contenedor):

```
# Docker
docker run -p 4741:4741 -p 4742:4742 -e KIVI_TOKEN=change-me iwasoftcom/kivi:1.1.0

# o el binario
KIVI_DATA=/var/lib/kivi/kivi.ledger KIVI_TOKEN=change-me kivid
```

Anexe un evento y léalo de vuelta — la respuesta llega con su traza:

```
kivi append ./kivi.ledger property '{"subject":"dog","attribute":"sound","value":"woof"}'
kivi table  ./kivi.ledger dog sound
# → {"scope":0,"trace":[0],"value":"woof"}

kivi verify ./kivi.ledger        # rehashear la cadena, verificar cada sello
kivi why    ./kivi.ledger 0      # el recibo detrás de la respuesta
```

Los clientes hablan el mismo contrato en **seis lenguajes** — Go, Python, Java & Kotlin, Rust, Node.js y .NET — con acceso tipado «entidad entra, entidad sale» y verificación del lado del cliente activada por defecto.

## Qué incluye

**Respuestas trazadas**  
Cada valor cita los eventos que lo establecen; una respuesta sin traza no puede construirse.

**Viaje en el tiempo**  
«¿Qué sabíamos en el registro N?» es un único replay acotado — sin instantáneas, sin migraciones (`--as-of`).

**Integridad verificada**  
Cadena de hash SHA-256 + sellos Ed25519; un solo byte alterado se detecta y se nombra.

**Borrado cripto**  
Claves por registro; borrar destruye la clave, mantiene la cadena válida y registra el borrado mismo.

**ACID, forma append-only**  
Atomicidad por evento, escrituras serializadas, lecturas por instantánea, durabilidad por group-commit `fsync`.

**Clúster & federación**  
Conmutación por error con commit mayoritario, réplicas de lectura y testimonio mutuo entre registros independientes.

**Panel de admin & identidad**  
UI integrada, configuración registrada, usuarios/roles/sesiones y claves de API con nombre y revocables.

**Puerta LLM (MCP)**  
Da a un agente de IA una memoria que cita sus fuentes; un hecho ausente se rechaza, no se alucina.

**Multiinquilino, bajo demanda**  
Varios inquilinos aislados en un proceso — y uno nuevo se abre en tiempo de ejecución vía la API de administración, con recibo y sin reinicio.

**Se ejecuta como servicio**  
systemd en Linux (.deb/.rpm) y un servicio real de Windows desde el instalador — arranque al inicio, drenaje limpio al parar.

## Panel de administración

Cada servidor incluye un panel web integrado (sin despliegue aparte) en el puerto de ops. Abra `https://<host>:4742/admin/` e inicie sesión con una cuenta de usuario — desde ahí, en vivo y sin tocar el plano de datos, puede:

-   **Resumen & Periodos** — salud, recuentos de registros y sellos, rotación.
-   **Registros & Búsqueda** — un flujo en vivo de los anexados, más búsqueda semántica.
-   **Clúster & Federación** — topología, estado de conmutación y de testimonio.
-   **Usuarios & Claves de API** — crear cuentas y roles, emitir y revocar claves.
-   **Configuración** — reajustar límites y ajustes; cada cambio se registra en el registro (los secretos nunca).

## Arquitectura

Un proceso, dos puertas de red, una regla de dependencia estricta: el núcleo usa solo la biblioteca estándar de Go; gRPC vive en el borde. En el disco hay exactamente un tipo de verdad duradera — el registro de eventos — y todo lo demás es descartable.

CLI · SDK en seis lenguajes desconfiado: verifica en el cliente Panel de admin · agentes LLM (MCP) UI integrada · memoria que cita plano de datos gRPC · :4741 ops HTTP · :4742 núcleo (Go stdlib): registro · vistas · traza · sello · identidad respuestas compiladas al consultar — nada derivado es primario disco: registro de eventos append-only (+ sellos)

## SDK cliente

Seis SDK cliente hablan el mismo contrato de cable — todos con licencia MIT y publicados en el registro de su lenguaje. El servidor y el núcleo de kivi siguen siendo propietarios; los clientes son abiertos.

| Lenguaje | Registro | Instalación | Página del paquete |
| --- | --- | --- | --- |
| Rust | crates.io | `cargo add kivi-sdk` | [crates.io/crates/kivi-sdk](https://crates.io/crates/kivi-sdk) |
| Python | PyPI | `pip install kivi-sdk` | [pypi.org/project/kivi-sdk](https://pypi.org/project/kivi-sdk/) |
| Node.js | npm | `npm install @iwasoft/kivi` | [npmjs.com/package/@iwasoft/kivi](https://www.npmjs.com/package/@iwasoft/kivi) |
| Java / Kotlin | Maven Central | `com.iwasoft:kivi:1.1.0` | [central.sonatype.com/artifact/com.iwasoft/kivi](https://central.sonatype.com/artifact/com.iwasoft/kivi) |
| .NET | NuGet | `dotnet add package Iwasoft.Kivi` | [nuget.org/packages/Iwasoft.Kivi](https://www.nuget.org/packages/Iwasoft.Kivi) |
| Go | GitHub | `go get github.com/iwasoftcom/kivi-sdk` | [github.com/iwasoftcom/kivi-sdk](https://github.com/iwasoftcom/kivi-sdk) |

### El cliente desconfiado

Cada SDK expone la misma superficie pequeña y tipada — y la verificación del lado del cliente está **activada por defecto**: el cliente vuelve a calcular el hash de la cadena y comprueba los sellos Ed25519 a medida que lee, de modo que un servidor mentiroso o un solo byte alterado se detecta, no se confía en él.

-   `append(type, body)` — añade un evento, obtén su recibo.
-   `table(subject, attribute)` — una **respuesta trazada**: `value` + `trace` (los números de registro que la establecieron) + `scope`. Una celda ausente es un rechazo honesto — una excepción nativa, nunca un null inventado.
-   `table(…, as_of=N)` — la misma respuesta **a fecha del registro N** (viaje en el tiempo).
-   `why(trace)` — los propios registros de recibo.
-   `replay()` — transmite cada registro, verificado de nuevo en el lado del cliente.
-   `similar(query, k)` — búsqueda semántica trazada (registro + puntuación).
-   `login(user, pw)` — un token de sesión acotado por rol; `with_bearer` reutiliza el canal bajo otra identidad.
-   `head()` — orientación barata (número de cabecera + hash), sin auditoría.

### Ejemplo (Python)

La forma es idéntica en los seis lenguajes:

```
from kivi import KiviClient

c = KiviClient("localhost:4741", token="…")     # verify=True by default
c.append("property", {"subject": "dog", "attribute": "sound", "value": "bark"})
a = c.table("dog", "sound")              # TracedAnswer(value="bark", trace=[0], scope=0)
old = c.table("dog", "sound", as_of=41)  # the answer as of record 41
receipts = c.why(a.trace)                # the actual ledger records
for rec in c.replay():                   # hash + chain + seal verified CLIENT-SIDE
    ...
```

Un solo examen de conformidad se ejecuta contra los seis, de modo que cada lenguaje devuelve la misma respuesta con la misma traza. Los ejemplos por lenguaje (Go, Java / Kotlin, Rust, Node.js, .NET) y la capa de entidades tipada están en la [referencia completa](reference.en.html).

## Documentación completa & código fuente

-   **[Referencia completa (inglés)](reference.en.html)** — cada función, la API de CLI y gRPC, las variables de entorno, las guías prácticas y ejemplos de código con pestañas por lenguaje.
-   **Compatibilidad:** la API gRPC, el formato de disco v2 y la superficie del SDK son un contrato SemVer — no se rompen sin un salto de versión mayor.
-   **Estado honesto:** aún sin auditoría de seguridad independiente; aún sin kilometraje en producción. Son divulgaciones, no reservas sobre la promesa de estabilidad.

## Contacto

Preguntas, una demo o una conversación como socio de diseño — escríbenos:

-   **Correo** — [info@iwasoft.com](mailto:info@iwasoft.com)
-   **LinkedIn** — [linkedin.com/company/iwasoft](https://www.linkedin.com/company/iwasoft)

[Contacto](#contact) · Kivi v1.1.0 · formato de disco v2 (contrato congelado, golden vectors) · núcleo en Go stdlib, gRPC en los bordes · un binario estático. © iwasoft.
