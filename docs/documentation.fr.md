Kivi

la base de données qui montre ses reçus · v1.1.0 · format v2

Une base de données à registre d'événements en Go pur — un seul binaire statique capable de rendre compte de chaque réponse qu'il donne.

## Ce qu'est Kivi

Kivi ne stocke que des **événements** — des enregistrements immuables, chaînés par hachage. Tout ce qui ressemble à un état courant (une table, un graphe, une série temporelle, un index vectoriel) est une **vue** : un pliage sur ces événements, recompilé à la demande et jetable à tout moment. Chaque réponse porte une **trace** (les numéros d'enregistrement qui l'établissent) et une **portée** (jusqu'où elle a regardé dans l'historique), et `why` renvoie les enregistrements-reçus eux-mêmes. Une donnée absente est un refus honnête, jamais une invention.

## Les trois principes

**Seuls les événements sont durables.**  
Il n'y a pas d'UPDATE. Les faits nouveaux sont ajoutés ; le passé n'est jamais réécrit. L'historique est de première classe.

**Les représentations sont compilées.**  
Supprimer `derived/` est toujours légitime : les réponses ne changent pas, seule la requête suivante est plus lente.

**Pas de réponse sans trace.**  
Une réponse sans trace est irreprésentable — dans le cœur et dans chaque SDK. Rien n'est inventé.

## En quoi il diffère d'une base de données classique

|  | Une base de données typique | Kivi |
| --- | --- | --- |
| Ce qui est durable | L'état courant ; UPDATE détruit le passé | Les événements ; l'état en est recompilé |
| Réponses | Valeurs nues | Valeur + trace + portée ; `why` récupère les reçus |
| Point dans le temps | Snapshots et machinerie supplémentaire | Gratuit par conception : plier jusqu'à l'enregistrement N (`--as-of`) |
| Intégrité | Supposée | Vérifiée : chaîne SHA-256 + sceaux Ed25519 ; les clients re-vérifient le serveur |
| Suppression | Ligne partie, histoire partie | Effacement crypto : clé détruite, octets inchangés, effacement acté par reçu |

## Démarrage rapide

Lancez le serveur (binaire statique unique ou image conteneur) :

```
# Docker
docker run -p 4741:4741 -p 4742:4742 -e KIVI_TOKEN=change-me iwasoftcom/kivi:1.1.0

# ou le binaire
KIVI_DATA=/var/lib/kivi/kivi.ledger KIVI_TOKEN=change-me kivid
```

Ajoutez un événement et relisez-le — la réponse arrive avec sa trace :

```
kivi append ./kivi.ledger property '{"subject":"dog","attribute":"sound","value":"woof"}'
kivi table  ./kivi.ledger dog sound
# → {"scope":0,"trace":[0],"value":"woof"}

kivi verify ./kivi.ledger        # re-hacher la chaîne, vérifier chaque sceau
kivi why    ./kivi.ledger 0      # le reçu derrière la réponse
```

Les clients parlent le même contrat en **six langages** — Go, Python, Java & Kotlin, Rust, Node.js et .NET — avec un accès typé « entité en entrée, entité en sortie » et la vérification côté client activée par défaut.

## Ce qu'il contient

**Réponses tracées**  
Chaque valeur cite les événements qui l'établissent ; une réponse sans trace ne peut être construite.

**Voyage temporel**  
« Que savions-nous à l'enregistrement N ? » est un unique replay borné — pas de snapshots, pas de migrations (`--as-of`).

**Intégrité vérifiée**  
Chaîne de hachage SHA-256 + sceaux Ed25519 ; un seul octet altéré est détecté et nommé.

**Effacement crypto**  
Clés par enregistrement ; effacer détruit la clé, garde la chaîne valide et acte l'effacement lui-même.

**ACID, forme append-only**  
Atomicité par événement, écritures sérialisées, lectures sur instantané, durabilité par group-commit `fsync`.

**Cluster & fédération**  
Bascule avec commit majoritaire, réplicas de lecture et témoignage mutuel entre registres indépendants.

**Panneau d'admin & identité**  
UI intégrée, configuration actée par reçu, utilisateurs/rôles/sessions et clés d'API nommées et révocables.

**Porte LLM (MCP)**  
Donne à un agent IA une mémoire qui cite ses sources ; un fait manquant est refusé, pas halluciné.

**Multi-tenant, à la demande**  
Plusieurs locataires isolés dans un seul processus — et un nouveau s'ouvre à l'exécution via l'API d'admin, acté par reçu, sans redémarrage.

**Fonctionne en service**  
systemd sous Linux (.deb/.rpm) et un vrai service Windows depuis l'installeur — démarrage au boot, vidange propre à l'arrêt.

## Panneau d'administration

Chaque serveur embarque un panneau web (aucun déploiement séparé) sur le port d'ops. Ouvrez `https://<host>:4742/admin/` et connectez-vous avec un compte utilisateur — de là, en direct et sans toucher au plan de données, vous pouvez :

-   **Vue d'ensemble & Périodes** — santé, nombre d'enregistrements et de sceaux, rotation.
-   **Enregistrements & Recherche** — un flux en direct des ajouts, plus la recherche sémantique.
-   **Cluster & Fédération** — topologie, état de bascule et de témoignage.
-   **Utilisateurs & Clés d'API** — créer comptes et rôles, émettre et révoquer des clés.
-   **Configuration** — réajuster limites et réglages ; chaque changement est acté par reçu dans le registre (jamais les secrets).

## Architecture

Un processus, deux portes réseau, une règle de dépendance stricte : le cœur n'utilise que la bibliothèque standard de Go ; gRPC vit à la périphérie. Sur le disque, il existe exactement un type de vérité durable — le journal d'événements — et tout le reste est jetable.

CLI · SDK en six langages méfiant : vérifie côté client Panneau d'admin · agents LLM (MCP) UI intégrée · mémoire qui cite plan de données gRPC · :4741 ops HTTP · :4742 cœur (Go stdlib) : registre · vues · trace · sceau · identité réponses compilées à la requête — rien de dérivé n'est primaire disque : journal d'événements append-only (+ sceaux)

## SDK client

Six SDK client parlent le même contrat de communication — tous sous licence MIT et publiés dans le registre de leur langage. Le serveur kivi et son cœur restent propriétaires ; les clients sont ouverts.

| Langage | Registre | Installation | Page du paquet |
| --- | --- | --- | --- |
| Rust | crates.io | `cargo add kivi-sdk` | [crates.io/crates/kivi-sdk](https://crates.io/crates/kivi-sdk) |
| Python | PyPI | `pip install kivi-sdk` | [pypi.org/project/kivi-sdk](https://pypi.org/project/kivi-sdk/) |
| Node.js | npm | `npm install @iwasoft/kivi` | [npmjs.com/package/@iwasoft/kivi](https://www.npmjs.com/package/@iwasoft/kivi) |
| Java / Kotlin | Maven Central | `com.iwasoft:kivi:1.1.0` | [central.sonatype.com/artifact/com.iwasoft/kivi](https://central.sonatype.com/artifact/com.iwasoft/kivi) |
| .NET | NuGet | `dotnet add package Iwasoft.Kivi` | [nuget.org/packages/Iwasoft.Kivi](https://www.nuget.org/packages/Iwasoft.Kivi) |
| Go | GitHub | `go get github.com/iwasoftcom/kivi-sdk` | [github.com/iwasoftcom/kivi-sdk](https://github.com/iwasoftcom/kivi-sdk) |

### Le client méfiant

Chaque SDK expose la même surface typée et réduite — et la vérification côté client est **activée par défaut** : le client recalcule le hachage de la chaîne et contrôle les sceaux Ed25519 au fil de la lecture, si bien qu'un serveur qui ment ou un seul octet altéré est détecté, non pas cru sur parole.

-   `append(type, body)` — ajoute un événement, obtient son reçu.
-   `table(subject, attribute)` — une **réponse tracée** : `value` + `trace` (les numéros d'enregistrement qui l'ont établie) + `scope`. Une cellule absente est un refus honnête — une exception native, jamais un null fabriqué.
-   `table(…, as_of=N)` — la même réponse **à la date de l'enregistrement N** (voyage temporel).
-   `why(trace)` — les enregistrements de reçu eux-mêmes.
-   `replay()` — diffuse chaque enregistrement, revérifié côté client.
-   `similar(query, k)` — recherche sémantique tracée (enregistrement + score).
-   `login(user, pw)` — un jeton de session limité à un rôle ; `with_bearer` réutilise le canal sous une autre identité.
-   `head()` — orientation économique (numéro de tête + hachage), sans audit.

### Exemple (Python)

La forme est identique dans les six langages :

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

Un seul examen de conformité s'exécute contre les six, de sorte que chaque langage renvoie la même réponse avec la même trace. Les exemples par langage (Go, Java / Kotlin, Rust, Node.js, .NET) et la couche d'entités typées se trouvent dans la [référence complète (en anglais)](reference.en.html).

## Documentation complète & code source

-   **[Référence complète (anglais)](reference.en.html)** — chaque fonctionnalité, l'API CLI et gRPC, les variables d'environnement, les guides pratiques et des exemples de code à onglets par langage.
-   **Compatibilité :** l'API gRPC, le format disque v2 et la surface SDK forment un contrat SemVer — ils ne cassent pas sans un saut de version majeure.
-   **Statut honnête :** pas encore audité en sécurité de façon indépendante ; pas encore de kilométrage en production. Ce sont des divulgations, pas des réserves sur la promesse de stabilité.

## Contact

Questions, une démo ou une discussion comme partenaire de conception — contactez-nous :

-   **E-mail** — [info@iwasoft.com](mailto:info@iwasoft.com)
-   **LinkedIn** — [linkedin.com/company/iwasoft](https://www.linkedin.com/company/iwasoft)

[Contact](#contact) · Kivi v1.1.0 · format disque v2 (contrat gelé, golden vectors) · cœur en Go stdlib, gRPC en périphérie · un binaire statique. © iwasoft.
