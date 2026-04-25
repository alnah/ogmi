# Descriptor specs

Ogmi reads descriptor specs from this directory by default, unless users provide
an external specs directory with `--specs` or `OGMI_SPECS`.

## Rights scope

The Apache License, Version 2.0 covers Ogmi software source code. It does not
relicense third-party descriptor text or source publications that may appear in
or inform descriptor specs.

Descriptor specs are not software source code and may have separate rights. Do
not assume that every file under `specs/` can be reused under the Apache License.

## Third-party material

Some current specs are under rights review because they include or derive from
third-party language-learning publications, including CEFR / CECRL material and
French linguistic inventory material. Third-party material remains subject to
its original rights.

Known source publications under review include:

- Council of Europe (2020), _Common European Framework of Reference for
  Languages: Learning, teaching, assessment - Companion volume_, Council of
  Europe Publishing, Strasbourg, available at <https://www.coe.int/lang-cefr>.
- _Inventaire linguistique des contenus clés des niveaux du CECRL_, Copyright
  Eaquals 2015, published with CIEP.

Permission requests are in progress. Until permissions are clear, do not treat
affected specs as reusable under the Ogmi software license.

## Original specs

New public Ogmi specs should be original. They may follow professional FLE
practice and general CEFR concepts, but they should not copy, translate, or
closely paraphrase third-party descriptor text or reproduce a third-party
publication's detailed structure.

## External specs

For private, licensed, experimental, or institution-specific descriptor data,
prefer an external specs directory:

```sh
ogmi --specs ./path/to/specs descriptors list
```

External specs let users keep rights-restricted or custom data outside the
public Ogmi repository and release artifacts.
