#!/usr/bin/env python3
"""Rebuild the attributed, ranked catalog from pinned local source files.

Python standard library only; never downloads data or rewrites authored lessons.
See SOURCE.md. Both input checksums are mandatory. Output is deterministic.
"""
from pathlib import Path
import argparse, collections, gzip, hashlib, json, re, uuid, zipfile
ROOT = Path(__file__).resolve().parents[4]
DEST = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('wordnet_zip', type=Path)
parser.add_argument('frequency_file', type=Path)
parser.add_argument('--output', type=Path, default=DEST/'catalog.json.gz')
args = parser.parse_args()
FREQ_URL = 'https://raw.githubusercontent.com/hermitdave/FrequencyWords/072bbed282316a23651aa7068c7173aa7898cf80/content/2018/en/en_50k.txt'
FREQ_SHA = '5351ff405b1126ef555791dd4d9798a48e3e9a501a9fc481a9da957752cfb458'
WORDNET_SHA = 'cbda5ea6eef7f36a97a43d4a75f85e07fccbb4f23657d27b4ccbc93e2646ab59'
raw = args.frequency_file.read_bytes()
if hashlib.sha256(raw).hexdigest() != FREQ_SHA:
    raise ValueError('Frequency source checksum does not match reviewed source')
original = args.wordnet_zip.read_bytes()
if hashlib.sha256(original).hexdigest() != WORDNET_SHA:
    raise ValueError('WordNet source checksum does not match reviewed source')
rows = [line.rsplit(' ', 1) for line in raw.decode().splitlines()]
archive = zipfile.ZipFile(args.wordnet_zip)
OVERRIDES = json.loads((DEST/'editorial-core.json').read_text())
EXCLUDE = set(json.loads((DEST/'excluded-entries.json').read_text()))
BLOCKED = EXCLUDE
records={}; parents={}; candidates=collections.defaultdict(list)
for kind,label in [('noun','noun'),('verb','verb'),('adj','adjective'),('adv','adverb')]:
 for line in archive.read('wordnet/data.'+kind).decode().splitlines():
  if not line or not line[0].isdigit():continue
  meta,gloss=line.split('|',1);f=meta.split();n=int(f[3],16)
  lemmas=[re.sub(r'\((?:a|p|ip)\)$','',f[4+2*i]) for i in range(n)]
  records[(kind,f[0])]=(lemmas,gloss.strip())
  j=4+2*n; pointer_count=int(f[j]); j+=1; broader=[]
  for _ in range(pointer_count):
   symbol,offset,pos,lexical=f[j:j+4];j+=4
   if symbol in ('@','@i'):
    broader.append({'n':'noun','v':'verb','a':'adj','s':'adj','r':'adv'}[pos]+':'+offset)
  parents[kind+':'+f[0]]=broader
 for line in archive.read('wordnet/index.sense').decode().splitlines():
  key,offset,num,count=line.split();lemma,keypart=key.split('%',1);ss=keypart[0]
  if ss not in {'noun':'1','verb':'2','adj':'35','adv':'4'}[kind]:continue
  if not re.fullmatch('[a-z]{2,}',lemma):continue
  record=records.get((kind,offset))
  if record is None:continue
  lemmas,gloss=record
  if lemma not in lemmas:continue
  candidates[lemma].append({'pos':label,'offset':offset,'sense_key':key,'tag_count':int(count),'sense_number':int(num),'definition':gloss,'example':''})
capitalized=set()
for lemmas,gloss in records.values():
 for lemma in lemmas:
  if lemma != lemma.lower():capitalized.add(lemma.lower())
verb_exceptions={}
for line in archive.read('wordnet/verb.exc').decode().splitlines():
 f=line.split();verb_exceptions[f[0]]=f[1:]
def is_inflection(word):
 bases=verb_exceptions.get(word,[])
 for suffix,replacement in [('ing',''),('ing','e'),('ed',''),('ed','e'),('ied','y'),('ies','y'),('es',''),('s','')]:
  if word.endswith(suffix):bases.append(word[:-len(suffix)]+replacement)
 return any(base != word and any(x['pos']=='verb' for x in candidates.get(base,[])) for base in bases)
seed=json.loads((ROOT/'apps/api/cmd/seed/voc026-p1.json').read_text()); existing={w['normalized_text']:w for w in seed['canonical_words']}; meanings=collections.defaultdict(list)
for m in seed['word_meanings']:meanings[m['word_id']].append(m)
ns=uuid.uuid5(uuid.NAMESPACE_URL,'https://vocanova.site/catalog/wordnet-3.0')
items=[]
for source_rank,(word,count) in enumerate(rows,1):
 if word in EXCLUDE or word in BLOCKED:continue
 if re.fullmatch(r'(?:goddamn(?:ed|it)?|piss(?:es|ed|ing)?|fuck(?:s|ed|er|ers|ing)?|motherfuck(?:er|ers|ing)?|shit(?:s|ted|ting|ty|head)?|arseholes?|assholes?|bullshit(?:s|ted|ting)?|damn(?:s|ed|ing)?)',word):continue
 if word in OVERRIDES:sense=dict(OVERRIDES[word]);source='Vocanova editorial core'
 elif word in candidates:
  # Tagged use from the historical WordNet concordance, not subtitle sense frequency.
  sense=dict(sorted(candidates[word],key=lambda x:(-x['tag_count'],x['sense_number'],x['sense_key']))[0]);source='WordNet 3.0'
 else:continue
 if source=='WordNet 3.0' and word not in existing:
  if sense['tag_count']<1:continue
  if word in capitalized and sense['tag_count']<10:continue
  if is_inflection(word) and sense['tag_count']<10:continue
  if re.search(r'(?i)ethnic slur|racial slur|offensive term|obscene|vulgar|derogatory|sexual intercourse|genital|penis|prostitute|archaic|obsolete|oriental|savage|retardation|intelligence quotient|\d',sense['definition']):continue
 if word in existing:
  old=existing[word];m=sorted(meanings[old['id']],key=lambda x:x['meaning_order'])[0]
  sense.update({'pos':m['part_of_speech'],'definition':m['short_definition'],'sense_key':'existing:'+m['id'],'example':''});source='Vocanova authored'
  wid=old['id'];mid=m['id']
 else:wid=str(uuid.uuid5(ns,'word:en:'+word));mid=str(uuid.uuid5(ns,'primary-meaning:en:'+word))
 gloss=sense['definition'];example=''
 match=re.search(r'; "([^"\n]+)"(?:; "[^"\n]+")*$',gloss)
 if match:example=match.group(1);gloss=gloss[:match.start()].rstrip()
 if source=='WordNet 3.0':
  # Source quotations can be unsafe, dated, or about a different synonym.
  # Retain the reference definition, never display unreviewed examples.
  gloss=gloss.split('; "',1)[0].rstrip()
  example=''
 items.append({'rank':len(items)+1,'source_rank':source_rank,'count':int(count),'word':word,'word_id':wid,'meaning_id':mid,'part_of_speech':sense['pos'],'definition':gloss,'example':example,'source':source,'source_sense_key':sense['sense_key'],'tag_count':sense.get('tag_count')})
 if len(items)==10000:break

assert len(items) == 10000, 'Filters no longer produce the complete reviewed inventory'
assert [x['rank'] for x in items] == list(range(1, 10001))
assert len({x['word'] for x in items}) == len(items)
assert len({x['word_id'] for x in items}) == len(items)
assert len({x['meaning_id'] for x in items}) == len(items)
payload = {
    'version': 'english-10000-v1',
    'frequency_source': {'url': FREQ_URL, 'sha256': FREQ_SHA,
        'author': 'Hermit Dave / FrequencyWords; OpenSubtitles 2018',
        'license': 'CC BY-SA 4.0', 'license_url': 'https://creativecommons.org/licenses/by-sa/4.0/',
        'changes': 'Filtered, ranked and joined with attributed WordNet senses and an original editorial core.'},
    'wordnet_source': {'url': 'https://github.com/nltk/nltk_data/blob/gh-pages/packages/corpora/wordnet.zip',
        'sha256': WORDNET_SHA, 'license_url': 'https://wordnet.princeton.edu/license-and-commercial-use'},
    'wordnet_license': archive.read('wordnet/LICENSE').decode('utf-8'),
    'items': items,
}
encoded = json.dumps(payload, ensure_ascii=False, sort_keys=True, separators=(',', ':')).encode()
packed = gzip.compress(encoded, compresslevel=9, mtime=0)
args.output.write_bytes(packed)
print(json.dumps({'words': len(items), 'json_bytes': len(encoded), 'gzip_bytes': len(packed),
    'sha256': hashlib.sha256(packed).hexdigest(), 'last_source_rank': items[-1]['source_rank']}))

# Semantic alias membership is independent of editorial selection and rank.
# Include every original sense for each current catalog/authored word, not only
# the chosen teaching sense. False-positive exclusions are safer than synonyms
# accidentally becoming wrong multiple-choice answers.
members = collections.defaultdict(set)
for (kind, offset), (lemmas, _) in records.items():
    for lemma in lemmas:
        members[lemma.lower().replace('_', ' ')].add(kind + ':' + offset)
exceptions = collections.defaultdict(set)
for kind in ['noun', 'verb', 'adj', 'adv']:
    for line in archive.read('wordnet/' + kind + '.exc').decode().splitlines():
        forms = line.split()
        for base in forms[1:]:
            exceptions[forms[0].replace('_', ' ')].add((kind, base.replace('_', ' ')))
suffixes = {
    'noun': [('s',''), ('ses','s'), ('xes','x'), ('zes','z'), ('ches','ch'), ('shes','sh'), ('men','man'), ('ies','y')],
    'verb': [('s',''), ('ies','y'), ('es','e'), ('es',''), ('ed','e'), ('ed',''), ('ing','e'), ('ing','')],
    'adj': [('er',''), ('est',''), ('er','e'), ('est','e')],
    'adv': [],
}
def sense_membership(word):
    found = set()
    for form in {word, word.replace('-', ' ')}:
        found.update(members.get(form, set()))
        bases = set(exceptions.get(form, set()))
        for kind, rules in suffixes.items():
            for suffix, replacement in rules:
                if form.endswith(suffix):
                    bases.add((kind, form[:-len(suffix)] + replacement))
        for kind, base in bases:
            found.update(synset for synset in members.get(base, set()) if synset.startswith(kind + ':'))
    return found
aliases = {word:sorted(sense_membership(word)) for word in sorted(set(existing) | {item['word'] for item in items})}
# Include lexical spellings of these concepts too (e.g. co-worker, booking),
# so normalization does not depend on whether a synonym itself is teachable.
active_senses = set().union(*(set(values) for values in aliases.values()))
for lemma, values in members.items():
    if values & active_senses:
        aliases.setdefault(lemma, sorted(sense_membership(lemma)))
# Reject direct broader/narrower meanings within two links. Do not compare
# shared ancestors: generic parents such as entity would remove unrelated words.
ancestors={}
for synset in sorted({sense for values in aliases.values() for sense in values}):
    direct=set(parents.get(synset,[]))
    broader=direct | {parent for middle in direct for parent in parents.get(middle,[])}
    if broader: ancestors[synset]=sorted(broader)
alias_payload = {'version':'wordnet-3.0-all-senses-v2', 'wordnet_license':payload['wordnet_license'], 'entries':aliases, 'ancestors':ancestors}
alias_raw = json.dumps(alias_payload, ensure_ascii=False, sort_keys=True, separators=(',', ':')).encode()
alias_packed = gzip.compress(alias_raw, compresslevel=9, mtime=0)
alias_output = args.output.with_name('semantic-aliases.json.gz')
alias_output.write_bytes(alias_packed)
print(json.dumps({'alias_words':len(aliases), 'alias_gzip_bytes':len(alias_packed), 'alias_sha256':hashlib.sha256(alias_packed).hexdigest()}))
