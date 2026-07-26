# kivi docs — translation glossary

Consistent terminology across all `documentation.<lang>.html` files. Code,
field names, event types (`kivi.*`), and product names stay in English in every
language. Where a term is conventionally kept in English in technical writing
(hash, append, commit, token…), keep it.

Source is English (`docs/kivi-docs.html`). Each translation is a copy of that
file with prose translated in place; CSS, code blocks, and SVG stay byte-identical.

| English | tr | de | fr | es | zh | ar |
|---|---|---|---|---|---|---|
| event ledger | olay defteri | Ereignis-Ledger | registre d'événements | registro de eventos | 事件账本 | سجل الأحداث |
| append-only | yalnızca-ekleme | Append-only | append-only | de solo anexado | 仅追加 | إلحاق فقط |
| record | kayıt | Datensatz | enregistrement | registro | 记录 | سِجِلّ |
| hash chain | özet zinciri | Hash-Kette | chaîne de hachage | cadena de hash | 哈希链 | سلسلة التجزئة |
| seal | mühür | Siegel | sceau | sello | 封印 | خَتم |
| trace | iz | Nachweis (trace) | trace | traza | 溯源 | أثر |
| receipt | fiş | Beleg | reçu | recibo | 回执 | إيصال |
| view (compiled) | görünüm | View (Sicht) | vue | vista | 视图 | عرض |
| fold | katlama | Faltung (fold) | pliage (fold) | plegado (fold) | 折叠 | طَيّ |
| derived (disposable) | türev | abgeleitet | dérivé | derivado | 派生 | مُشتَقّ |
| durable | kalıcı | dauerhaft | durable | duradero | 持久 | دائم |
| tamper | kurcalama | Manipulation | altération | manipulación | 篡改 | عَبَث |
| crypto-erase | kripto-silme | Krypto-Löschung | effacement crypto | borrado cripto | 加密擦除 | محو تشفيري |
| snapshot | fotoğraf/anlık | Snapshot | instantané | instantánea | 快照 | لقطة |
| as-of (time travel) | -anına (zaman yolculuğu) | Stand von (Zeitreise) | à la date (voyage temporel) | a fecha de (viaje en el tiempo) | 时间旅行 | حتى لحظة |
| offset index | ofset indeksi | Offset-Index | index de décalage | índice de desplazamiento | 偏移索引 | فهرس الإزاحة |
| replica / follower | kopya / takipçi | Replik / Follower | réplique / suiveur | réplica / seguidor | 副本 / 跟随者 | نسخة / تابع |
| cluster | küme | Cluster | cluster | clúster | 集群 | عنقود |
| federation | federasyon | Föderation | fédération | federación | 联邦 | اتحاد |
| conformance exam | konformans sınavı | Konformitätsprüfung | examen de conformité | examen de conformidad | 一致性测试 | اختبار المطابقة |
| honest refusal | dürüst ret | ehrliche Ablehnung | refus honnête | rechazo honesto | 诚实拒绝 | رفض صادق |
| ledger integrity | defter bütünlüğü | Ledger-Integrität | intégrité du registre | integridad del registro | 账本完整性 | سلامة السجل |

Notes:
- Product/proper names unchanged: **kivi, Kivi, gRPC, Ed25519, JSON, SDK, MCP, K8s, OpenShift, Docker, TLS**.
- Arabic (`ar`) files set `<html lang="ar" dir="rtl">`; `<pre>`/`<code>` blocks keep `dir="ltr"`.
- Chinese (`zh`) is LTR.
