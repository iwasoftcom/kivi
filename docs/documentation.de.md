Kivi

die Datenbank, die ihre Belege zeigt · v1.1.0 · Format v2

Eine Event-Ledger-Datenbank in reinem Go — eine einzige statische Binärdatei, die für jede ihrer Antworten Rechenschaft ablegen kann.

## Was Kivi ist

Kivi speichert nur **Ereignisse** — unveränderliche, hash-verkettete Datensätze. Alles, was wie aktueller Zustand aussieht (eine Tabelle, ein Graph, eine Zeitreihe, ein Vektorindex), ist eine **View**: eine Faltung über diese Ereignisse, bei Bedarf neu berechnet und jederzeit verwerfbar. Jede Antwort trägt einen **Nachweis** (die Datensatznummern, die sie begründen) und einen **Scope** (wie weit in die Historie geblickt wurde), und `why` liefert die Belegdatensätze selbst. Fehlende Daten sind eine ehrliche Ablehnung, niemals eine Erfindung.

## Die drei Prinzipien

**Nur Ereignisse sind dauerhaft.**  
Es gibt kein UPDATE. Neue Fakten werden angehängt; die Vergangenheit wird nie umgeschrieben. Historie ist erstklassig.

**Repräsentationen werden berechnet.**  
Das Löschen von `derived/` ist immer legitim: Antworten ändern sich nicht, nur die nächste Abfrage ist langsamer.

**Keine Antwort ohne Nachweis.**  
Eine nachweislose Antwort ist nicht darstellbar — im Kern und in jedem SDK. Nichts wird erfunden.

## Wie es sich von einer normalen Datenbank unterscheidet

|  | Eine typische Datenbank | Kivi |
| --- | --- | --- |
| Was dauerhaft ist | Aktueller Zustand; UPDATE zerstört die Vergangenheit | Die Ereignisse; der Zustand wird daraus neu berechnet |
| Antworten | Nackte Werte | Wert + Nachweis + Scope; `why` holt die Belege |
| Zeitpunkt-Abfrage | Snapshots und zusätzliche Maschinerie | Von Natur aus kostenlos: bis Datensatz N falten (`--as-of`) |
| Integrität | Angenommen | Verifiziert: SHA-256-Kette + Ed25519-Siegel; Clients prüfen den Server erneut |
| Löschen | Zeile weg, Geschichte weg | Krypto-Löschung: Schlüssel zerstört, Bytes unverändert, Löschung belegt |

## Schnellstart

Server starten (einzelne statische Binärdatei oder Container-Image):

```
# Docker
docker run -p 4741:4741 -p 4742:4742 -e KIVI_TOKEN=change-me iwasoftcom/kivi:1.1.0

# oder die Binärdatei
KIVI_DATA=/var/lib/kivi/kivi.ledger KIVI_TOKEN=change-me kivid
```

Ein Ereignis anhängen und zurücklesen — die Antwort kommt mit ihrem Nachweis:

```
kivi append ./kivi.ledger property '{"subject":"dog","attribute":"sound","value":"woof"}'
kivi table  ./kivi.ledger dog sound
# → {"scope":0,"trace":[0],"value":"woof"}

kivi verify ./kivi.ledger        # Kette neu hashen, jedes Siegel prüfen
kivi why    ./kivi.ledger 0      # der Beleg hinter der Antwort
```

Clients sprechen denselben Vertrag in **sechs Sprachen** — Go, Python, Java & Kotlin, Rust, Node.js und .NET — mit typisiertem Zugriff „Entity rein, Entity raus" und standardmäßig aktiver clientseitiger Verifikation.

## Was drin ist

**Nachvollziehbare Antworten**  
Jeder Wert nennt die Ereignisse, die ihn begründen; eine nachweislose Antwort ist nicht baubar.

**Zeitreise**  
„Was wussten wir bei Datensatz N?" ist ein begrenztes Replay — keine Snapshots, keine Migrationen (`--as-of`).

**Verifizierte Integrität**  
SHA-256-Hash-Kette + Ed25519-Siegel; ein einziges gekipptes Byte wird erkannt und benannt.

**Krypto-Löschung**  
Schlüssel pro Datensatz; Löschen zerstört den Schlüssel, hält die Kette gültig und belegt die Löschung selbst.

**ACID, Append-only-Form**  
Atomizität pro Ereignis, serialisierte Schreibvorgänge, Snapshot-Lesevorgänge, Group-Commit-`fsync`\-Dauerhaftigkeit.

**Cluster & Föderation**  
Failover mit Mehrheits-Commit, Read-Repliken und gegenseitige Bezeugung zwischen unabhängigen Ledgern.

**Admin-Panel & Identität**  
Eingebettete UI, belegte Konfiguration, Benutzer/Rollen/Sitzungen und benannte, widerrufbare API-Schlüssel.

**LLM-Tür (MCP)**  
Gibt einem KI-Agenten ein Gedächtnis, das seine Quellen nennt; ein fehlender Fakt wird abgelehnt, nicht halluziniert.

**Mandantenfähig, bei Bedarf**  
Mehrere isolierte Mandanten in einem Prozess — und ein neuer lässt sich zur Laufzeit über die Admin-API eröffnen, belegt, ohne Neustart.

**Läuft als Dienst**  
systemd unter Linux (.deb/.rpm) und ein echter Windows-Dienst aus dem Installer — Start beim Booten, sauberes Entleeren beim Stoppen.

## Admin-Panel

Jeder Server bringt ein eingebettetes Web-Panel mit (keine separate Bereitstellung), am Ops-Port. Öffnen Sie `https://<host>:4742/admin/` und melden Sie sich mit einem Benutzerkonto an — von dort aus können Sie live und ohne die Datenebene zu berühren:

-   **Übersicht & Perioden** — Zustand, Datensatz- und Siegelzahlen, Rotation.
-   **Datensätze & Suche** — ein Live-Stream der Appends, dazu semantische Suche.
-   **Cluster & Föderation** — Topologie, Failover- und Bezeugungsstatus.
-   **Benutzer & API-Schlüssel** — Konten und Rollen anlegen, Schlüssel ausstellen und widerrufen.
-   **Konfiguration** — Limits und Einstellungen nachjustieren; jede Änderung wird ins Ledger belegt (Geheimnisse nie).

## Architektur

Ein Prozess, zwei Netzwerktüren, eine strikte Abhängigkeitsregel: Der Kern nutzt nur die Go-Standardbibliothek; gRPC lebt am Rand. Auf der Festplatte gibt es genau eine Art dauerhafter Wahrheit — das Ereignisprotokoll — und alles andere ist verwerfbar.

CLI · SDKs in sechs Sprachen misstrauisch: clientseitig verifizieren Admin-Panel · LLM-Agenten (MCP) eingebettete UI · zitierendes Gedächtnis gRPC-Datenebene · :4741 Ops-HTTP · :4742 Kern (Go-stdlib): Ledger · Views · Nachweis · Siegel · Identität Antworten zur Abfragezeit berechnet — nichts Abgeleitetes ist primär Festplatte: Append-only-Ereignisprotokoll (+ Siegel)

## Client-SDKs

Sechs Client-SDKs sprechen denselben Wire-Vertrag — alle MIT-lizenziert und in der Registry ihrer jeweiligen Sprache veröffentlicht. Der kivi-Server und der Kern bleiben proprietär; die Clients sind offen.

| Sprache | Registry | Installation | Paketseite |
| --- | --- | --- | --- |
| Rust | crates.io | `cargo add kivi-sdk` | [crates.io/crates/kivi-sdk](https://crates.io/crates/kivi-sdk) |
| Python | PyPI | `pip install kivi-sdk` | [pypi.org/project/kivi-sdk](https://pypi.org/project/kivi-sdk/) |
| Node.js | npm | `npm install @iwasoft/kivi` | [npmjs.com/package/@iwasoft/kivi](https://www.npmjs.com/package/@iwasoft/kivi) |
| Java / Kotlin | Maven Central | `com.iwasoft:kivi:1.1.0` | [central.sonatype.com/artifact/com.iwasoft/kivi](https://central.sonatype.com/artifact/com.iwasoft/kivi) |
| .NET | NuGet | `dotnet add package Iwasoft.Kivi` | [nuget.org/packages/Iwasoft.Kivi](https://www.nuget.org/packages/Iwasoft.Kivi) |
| Go | GitHub | `go get github.com/iwasoftcom/kivi-sdk` | [github.com/iwasoftcom/kivi-sdk](https://github.com/iwasoftcom/kivi-sdk) |

### Der misstrauische Client

Jedes SDK bietet dieselbe kleine, typisierte Oberfläche — und die clientseitige Verifikation ist **standardmäßig aktiv**: Der Client hasht die Kette beim Lesen neu und prüft die Ed25519-Siegel, sodass ein lügender Server oder ein einziges gekipptes Byte erkannt und nicht vertraut wird.

-   `append(type, body)` — ein Ereignis hinzufügen, seinen Beleg erhalten.
-   `table(subject, attribute)` — eine **nachweisbelegte Antwort**: `value` + `trace` (die Datensatznummern, die sie begründet haben) + `scope`. Eine fehlende Zelle ist eine ehrliche Ablehnung — eine native Ausnahme, niemals ein erfundenes null.
-   `table(…, as_of=N)` — dieselbe Antwort **zum Stand von Datensatz N** (Zeitreise).
-   `why(trace)` — die Beleg-Datensätze selbst.
-   `replay()` — jeden Datensatz streamen, clientseitig neu verifiziert.
-   `similar(query, k)` — nachweisbelegte semantische Suche (Datensatz + Score).
-   `login(user, pw)` — ein rollenbeschränktes Sitzungs-Token; `with_bearer` nutzt den Kanal unter einer anderen Identität wieder.
-   `head()` — günstige Orientierung (Head-Nummer + Hash), keine Prüfung.

### Beispiel (Python)

Die Form ist in allen sechs Sprachen identisch:

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

Eine einzige Konformitätsprüfung läuft gegen alle sechs, sodass jede Sprache dieselbe Antwort mit demselben Nachweis (trace) zurückgibt. Sprachspezifische Beispiele (Go, Java / Kotlin, Rust, Node.js, .NET) und die typisierte Entity-Schicht finden Sie in der [vollständigen Referenz](reference.en.html).

## Vollständige Dokumentation & Quellcode

-   **[Vollständige Referenz (Englisch)](reference.en.html)** — jede Funktion, die CLI- und gRPC-API, Umgebungsvariablen, How-to-Anleitungen und Codebeispiele mit Sprach-Tabs.
-   **Kompatibilität:** die gRPC-API, das Festplattenformat v2 und die SDK-Oberfläche sind ein SemVer-Vertrag — sie brechen nicht ohne einen Major-Versionssprung.
-   **Ehrlicher Status:** noch nicht unabhängig sicherheitsauditiert; noch keine Produktionslaufleistung. Dies sind Offenlegungen, keine Vorbehalte am Stabilitätsversprechen.

## Kontakt

Fragen, eine Demo oder ein Gespräch als Design-Partner — erreichen Sie uns:

-   **E-Mail** — [info@iwasoft.com](mailto:info@iwasoft.com)
-   **LinkedIn** — [linkedin.com/company/iwasoft](https://www.linkedin.com/company/iwasoft)

[Kontakt](#contact) · Kivi v1.1.0 · Festplattenformat v2 (eingefrorener Vertrag, Golden Vectors) · Go-stdlib-Kern, gRPC an den Rändern · eine statische Binärdatei. © iwasoft.
