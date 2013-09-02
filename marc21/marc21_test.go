package marc21

/*
   Go Language MARC21 Library
   Copyright (C) 2011 William Waites

   This program is free software: you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public License
   as published by the Free Software Foundation, either version 3 of
   the License, or (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License and the GNU General Public License along with this program
   (the files COPYING and GPL3 respectively).  If not, see
   <http://www.gnu.org/licenses/>.
*/

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func openTestMARC(t *testing.T) (data *os.File) {
	data, err := os.Open("test.mrc")
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestReader(t *testing.T) {
	exp := 85

	data := openTestMARC(t)
	defer data.Close()

	count := 0
	for {
		_, err := ReadRecord(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		count++
		/*
			r.XML(os.Stdout)
			log.Print(record)
				buf := &bytes.Buffer{}
				err = record.XML(buf)
				if err != nil {
					log.Fatal(err)
				}
				log.Print(string(buf.Bytes()))
				break
		*/
	}
	if count != exp {
		t.Errorf("Expected to read %d records, got %d", exp, count)
	}
	log.Printf("%d records", count)
}

func TestString(t *testing.T) {
	const exp = `001 50001
005 20010903131819.0
008 701012s1970    moua     b    001 0 eng  
010 [  ] [(a)    73117956 ]
035 [  ] [(a) ocm00094426 ]
035 [  ] [(9) 7003024381]
040 [  ] [(a) DLC], [(c) DLC], [(d) OKO]
020 [  ] [(a) 0801657024]
050 [00] [(a) RC78.7.C9], [(b) Z83]
060 [  ] [(a) QS 504 Z94d 1970]
082 [00] [(a) 616.07/583]
049 [  ] [(a) CUDA]
100 [1 ] [(a) Zugibe, Frederick T.], [(q) (Frederick Thomas),], [(d) 1928-]
245 [10] [(a) Diagnostic histochemistry], [(c) [by] Frederick T. Zugibe.]
260 [  ] [(a) Saint Louis,], [(b) Mosby,], [(c) 1970.]
300 [  ] [(a) xiv, 366 p.], [(b) illus.], [(c) 25 cm.]
504 [  ] [(a) Bibliography: p. 332-349.]
650 [ 0] [(a) Cytodiagnosis.]
650 [ 0] [(a) Histochemistry], [(x) Technique.]
650 [ 2] [(a) Histocytochemistry.]
650 [ 2] [(a) Histological Techniques.]
994 [  ] [(a) 92], [(b) CUD]`

	data := openTestMARC(t)
	defer data.Close()

	r, _ := ReadRecord(data)
	out := r.String()
	if out != exp {
		t.Errorf("Output record string %s did not match expected string", out)
	}
	log.Printf("Record.String()")
}

func TestToXML(t *testing.T) {
	const exp = `<record><leader>00819cam a2200289   4500</leader><controlfield tag="001">50001</controlfield><controlfield tag="005">20010903131819.0</controlfield><controlfield tag="008">701012s1970    moua     b    001 0 eng  </controlfield><datafield tag="010" ind1="32" ind2="32"><subfield code="97">   73117956 </subfield></datafield><datafield tag="035" ind1="32" ind2="32"><subfield code="97">ocm00094426 </subfield></datafield><datafield tag="035" ind1="32" ind2="32"><subfield code="57">7003024381</subfield></datafield><datafield tag="040" ind1="32" ind2="32"><subfield code="97">DLC</subfield><subfield code="99">DLC</subfield><subfield code="100">OKO</subfield></datafield><datafield tag="020" ind1="32" ind2="32"><subfield code="97">0801657024</subfield></datafield><datafield tag="050" ind1="48" ind2="48"><subfield code="97">RC78.7.C9</subfield><subfield code="98">Z83</subfield></datafield><datafield tag="060" ind1="32" ind2="32"><subfield code="97">QS 504 Z94d 1970</subfield></datafield><datafield tag="082" ind1="48" ind2="48"><subfield code="97">616.07/583</subfield></datafield><datafield tag="049" ind1="32" ind2="32"><subfield code="97">CUDA</subfield></datafield><datafield tag="100" ind1="49" ind2="32"><subfield code="97">Zugibe, Frederick T.</subfield><subfield code="113">(Frederick Thomas),</subfield><subfield code="100">1928-</subfield></datafield><datafield tag="245" ind1="49" ind2="48"><subfield code="97">Diagnostic histochemistry</subfield><subfield code="99">[by] Frederick T. Zugibe.</subfield></datafield><datafield tag="260" ind1="32" ind2="32"><subfield code="97">Saint Louis,</subfield><subfield code="98">Mosby,</subfield><subfield code="99">1970.</subfield></datafield><datafield tag="300" ind1="32" ind2="32"><subfield code="97">xiv, 366 p.</subfield><subfield code="98">illus.</subfield><subfield code="99">25 cm.</subfield></datafield><datafield tag="504" ind1="32" ind2="32"><subfield code="97">Bibliography: p. 332-349.</subfield></datafield><datafield tag="650" ind1="32" ind2="48"><subfield code="97">Cytodiagnosis.</subfield></datafield><datafield tag="650" ind1="32" ind2="48"><subfield code="97">Histochemistry</subfield><subfield code="120">Technique.</subfield></datafield><datafield tag="650" ind1="32" ind2="50"><subfield code="97">Histocytochemistry.</subfield></datafield><datafield tag="650" ind1="32" ind2="50"><subfield code="97">Histological Techniques.</subfield></datafield><datafield tag="994" ind1="32" ind2="32"><subfield code="97">92</subfield><subfield code="98">CUD</subfield></datafield></record>`

	data := openTestMARC(t)
	defer data.Close()

	r, err := ReadRecord(data)
	buf := &bytes.Buffer{}
	err = r.XML(buf)
	if err != nil {
		t.Fatal(err)
	}
	if exp != string(buf.Bytes()) {
		t.Errorf("Output XML did not match expected XML")
	}
	log.Printf("XML output")
}

func TestGetFields(t *testing.T) {
	const exp = `Record: 1
650 [ 0] [(a) Cytodiagnosis.]
650 [ 0] [(a) Histochemistry], [(x) Technique.]
650 [ 2] [(a) Histocytochemistry.]
650 [ 2] [(a) Histological Techniques.]
Record: 2
650 [ 0] [(a) Logic, Symbolic and mathematical.]
Record: 4
650 [ 0] [(a) Landscape drawing.]
650 [ 0] [(a) Colors.]
Record: 5
650 [ 0] [(a) Musicals], [(z) United States], [(x) History and criticism.]
650 [ 0] [(a) Musicals], [(x) Discography.]
Record: 6
650 [ 0] [(a) Opioid abuse.]
Record: 7
650 [ 4] [(a) Taxation], [(z) England], [(y) 1991.]
650 [ 4] [(a) England], [(x) Taxation], [(y) 1991.]
650 [ 0] [(a) Taxation], [(z) Great Britain.]
Record: 8
650 [ 0] [(a) Dragons], [(z) China.]
650 [ 0] [(a) Dragons in art.]
650 [ 0] [(a) Art, Chinese.]
Record: 12
650 [ 0] [(a) Navies], [(x) Insignia.]
Record: 13
650 [ 0] [(a) Real property tax], [(z) England], [(x) History.]
Record: 14
650 [ 0] [(a) Nuclear reactors], [(x) Control.]
650 [ 0] [(a) Neutron flux.]
Record: 17
650 [ 0] [(a) Psychopharmacology.]
650 [ 0] [(a) Psychotropic drugs.]
Record: 18
650 [ 0] [(a) Advertising.]
Record: 19
650 [ 0] [(a) God], [(x) History of doctrines.]
650 [ 0] [(a) Secularization.]
Record: 20
650 [ 0] [(a) Germans], [(z) Czechoslovakia], [(x) History.]
650 [ 0] [(a) Minorities], [(z) Czechoslovakia], [(x) History.]
Record: 22
650 [ 0] [(a) Homeopathy], [(x) Materia medica and therapeutics.]
Record: 23
650 [ 0] [(a) Songs with piano.]
650 [ 0] [(a) Patriotic music.]
Record: 24
650 [ 0] [(a) Acoustical engineering.]
650 [ 0] [(a) Noise control.]
Record: 26
650 [ 0] [(a) Judaism], [(x) Liturgy.]
650 [ 0] [(a) Judaism], [(x) Customs and practices.]
650 [ 0] [(a) Fasts and feasts], [(x) Judaism.]
Record: 27
650 [ 0] [(a) Electronic monitoring of parolees and probationers.]
Record: 29
650 [ 0] [(a) Tourism.]
Record: 30
650 [ 0] [(a) Computers and civilization], [(v) Humor.]
Record: 31
650 [ 0] [(a) Authors, Slovak], [(v) Diaries.]
650 [ 0] [(a) World War, 1939-1945], [(x) Personal narratives, Slovak.]
Record: 32
650 [ 0] [(a) Latin language, Medieval and modern], [(x) Study and teaching], [(v) Congresses.]
Record: 33
650 [ 0] [(a) Missions], [(z) South Africa.]
650 [ 0] [(a) Zulu (African people)], [(x) Missions.]
Record: 35
650 [ 0] [(a) Novelists, English], [(y) 20th century], [(v) Biography.]
Record: 37
650 [ 0] [(a) Negotiation in business.]
Record: 38
650 [ 0] [(a) Land reform], [(z) Zimbabwe.]
650 [ 0] [(a) Land use], [(x) Government policy], [(z) Zimbabwe.]
650 [ 0] [(a) Democracy], [(z) Zimbabwe.]
Record: 39
650 [ 0] [(a) Prints], [(y) 19th century], [(z) England], [(v) Exhibitions.]
650 [ 0] [(a) Prints, English], [(v) Exhibitions.]
Record: 40
650 [ 0] [(a) Germans], [(z) Czechoslovakia], [(x) History.]
650 [ 0] [(a) Minorities], [(z) Czechoslovakia], [(x) History.]
Record: 42
650 [ 0] [(a) Upper class], [(z) England], [(v) Humor.]
650 [ 0] [(a) Etiquette], [(z) England], [(v) Humor.]
650 [ 0] [(a) Eccentrics and eccentricities], [(z) England], [(v) Humor.]
Record: 43
650 [ 0] [(a) Conduct of life.]
650 [ 0] [(a) Prostitution], [(z) Great Britain], [(x) History], [(y) 19th century], [(v) Sources.]
650 [ 0] [(a) Women], [(z) Great Britain], [(x) Social conditions.]
Record: 44
650 [ 0] [(a) Authors, English], [(y) 17th century], [(v) Biography.]
650 [ 0] [(a) Authors, English], [(y) 18th century], [(v) Biography.]
Record: 58
650 [ 0] [(a) Beauty operators], [(z) England], [(v) Biography.]
Record: 59
650 [ 0] [(a) Greek language, Modern.]
650 [ 0] [(a) Mythology, Greek.]
650 [ 0] [(a) Greek poetry, Modern], [(x) History and criticism.]
Record: 62
650 [ 0] [(a) Education], [(x) Philosophy.]
650 [ 0] [(a) Art], [(x) Philosophy.]
Record: 63
650 [ 0] [(a) Inland navigation], [(z) Balkan Peninsula.]
Record: 64
650 [ 0] [(a) Exhumation], [(z) Saint Helena.]
650 [ 0] [(a) Emperors], [(z) France], [(x) Death.]
Record: 65
650 [ 0] [(a) Catechisms, Greek.]
Record: 66
650 [ 0] [(a) History], [(x) Methodology.]
650 [ 0] [(a) Linguistics.]
Record: 68
650 [ 0] [(a) Clothing and dress], [(z) Japan], [(x) History.]
650 [ 0] [(a) Used clothing industry], [(z) Japan], [(x) History.]
Record: 70
650 [ 0] [(a) Painting, German.]
650 [ 0] [(a) Painting, Flemish.]
650 [ 0] [(a) Painting, Dutch.]
Record: 71
650 [ 0] [(a) Repudiation.]
650 [ 0] [(a) State bankruptcy.]
Record: 72
650 [ 0] [(a) Slavic philology.]
Record: 73
650 [ 0] [(a) Victims of state-sponsored terrorism], [(z) Northern Ireland.]
650 [ 0] [(a) Political violence], [(z) Northern Ireland.]
650 [ 0] [(a) Violent deaths], [(z) Northern Ireland.]
650 [ 0] [(a) Civil rights], [(z) Northern Ireland.]
Record: 74
650 [ 0] [(a) Slavery and the church], [(z) Southern States], [(x) History.]
650 [ 0] [(a) Slavery], [(x) Moral and ethical aspects], [(z) Southern States], [(x) History.]
650 [ 0] [(a) Slavery], [(z) Southern States], [(x) History.]
650 [ 0] [(a) Christianity and culture], [(z) Southern States], [(x) History.]
650 [ 0] [(a) Culture conflict], [(z) Southern States], [(x) History.]
Record: 75
650 [ 0] [(a) Songs with piano.]
Record: 78
650 [ 0] [(a) Symphonies], [(v) Excerpts], [(v) Scores.]
Record: 79
650 [ 0] [(a) Astronomy in literature.]
650 [ 0] [(a) Cosmology in literature.]
650 [ 0] [(a) Literature and science], [(x) History.]
Record: 80
650 [ 0] [(a) Sermons, Scottish], [(y) 19th century.]`

	data := openTestMARC(t)
	defer data.Close()

	count := 0
	subjects := []string{}
	for {
		r, err := ReadRecord(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		count++
		t := r.GetFields("650")
		if len(t) != 0 {
			subjects = append(subjects, fmt.Sprintf("Record: %d", count))
			for _, f := range t {
				subjects = append(subjects, fmt.Sprintf("%s", f))
			}
		}
	}
	out := strings.Join(subjects, "\n")
	if out != exp {
		t.Error("Returned fields did not match expected fields")
	}
	log.Printf("GetFields()")
}

func TestGetSubFields(t *testing.T) {
	const exp = `Record 1: (a) Diagnostic histochemistry
Record 2: (a) Elements of mathematical logic :
Record 3: (a) The Narren-motifs in the works of Georg Büchner.
Record 4: (a) The way to sketch,
Record 5: (a) The American musical theater;
Record 6: (a) Opioids :
Record 7: (a) Mayson on revenue law.
Record 8: (a) Céleste dragon :
Record 9: (a) The Geography of Lograire /
Record 10: (a) Forskarbiografin :
Record 11: (a) Swedish imprints 1731-1833 :
Record 12: (a) Naval and Marine badges and insignia of World War 2 /
Record 13: (a) Land tax assessments c.1690-c.1950 /
Record 14: (a) Upravlenie neĭtronnym polem i︠a︡dernogo reaktora /
Record 15: (a) The final analysis of Dr Stark /
Record 16: (a) Érintések /
Record 17: (a) Clinical pharmacology of psychotherapeutic drugs /
Record 18: (a) Behind the scenes in advertising /
Record 19: (a) The way of transcendence :
Record 20: (a) Integration oder Ausgrenzung :
Record 21: (a) Saunders Lewis /
Record 22: (a) A materia medica of homeopathic formulas.
Record 23: (a) Our monarch, the Prince and the nation :
Record 24: (a) Acoustics and noise control /
Record 25: (a) A letter to Dr. Bentley :
Record 26: (a) Sefer Moreh be-ʾeṣbaʻ /
Record 27: (a) Electronic tagging :
Record 28: (a) The ebbing of the kraft /
Record 29: (a) The business of tourism /
Record 30: (a) The devouring fungus :
Record 31: (a) Surová býva vše pravda života ...
Record 32: (a) Vocabulary of teaching and research between Middle Ages and Renaissance :
Record 33: (a) Umfundisi :
Record 34: (a) Country life, and Society and solitude :
Record 35: (a) My road :
Record 36: (a) Nienasycenie /
Record 37: (a) The outstanding negotiator :
Record 38: (a) Land and democracy in Zimbabwe /
Record 39: (a) Berufskünstler und Amateure, Whistler, Haden und die Blüte der Graphik in England :
Record 40: (a) Integration oder Ausgrenzung :
Record 41: (a) Saunders Lewis /
Record 42: (a) The English gentleman /
Record 43: (a) Social purity /
Record 44: (a) Lives of Dryden and Pope /
Record 45: (a) A system of medicine... Vol.5 /
Record 46: (a) Parley's present for all seasons /
Record 47: (a) An act for dividing and inclosing the open and common fields, meadows, and common fen within the parishes of Billingborough and Birthorpe, in the county of Lincoln, and for draining and improving the said fen.
Record 48: (a) [Versus Christianismus, edur Sannur Christen̄domur, i fiorum Bokum... Saman̄skrifadur af Johanne Arndt... En̄ nu...wtlagdur a Islensku af...Þorleife Arnaysyne.
Record 49: (a) Zhong hua da zang jing (Han wen bu fen).
Record 50: (a) Sôgyokushi :
Record 51: (a) Newsletter, An Foras Riarachain.
Record 52: (a) Coop Developer.
Record 53: (a) The Institute's professional qualifying examination and membership regulations /
Record 54: (a) [The New Testament].
Record 55: (a) In der Sprache der Sagas :
Record 56: (a) The book of Psalms :
Record 57: (a) Histoire sommaire de la civilisation depuis l'origine jusqu'à nos jours /
Record 58: (a) My life (as I remember it) /
Record 59: (a) Horae Hellenicæ :
Record 60: (a) The Dance of death /
Record 61: (a) Blaise Cendrars, une étude par Louis Parrot, un choix de poèmes et de textes, une bibliographie établie par J. H. Levesque, des inédits, des manuscrits, des dessins, des portraits.
Record 62: (a) Plato's ideas on art and education /
Record 63: (a) The Danube-Aegean waterway project :
Record 64: (a) Journal du retour des cendres, 1840 :
Record 65: (a) [Katēchēseis tēs Christianikēs pisteōs,] :
Record 66: (a) Histoire et linguistique.
Record 67: (a) The infallibility of the church. A course of lectures delivered in the Divinity School of the University of Dublin.
Record 68: (a) Furugi /
Record 69: (a) The Royal River: the Thames, from source to sea. Descriptive, historical, pictorial /
Record 70: (a) Handbook of painting :
Record 71: (a) The repudiation of state debts :
Record 72: (a) Slavi︠a︡nskai︠a︡ filologii︠a︡:
Record 73: (a) Unfinished business :
Record 74: (a) Black & tan :
Record 75: (a) Fem sang =
Record 76: (a) The Church of St. Edward, King and Martyr, Corfe Castle.
Record 77: (a) Four Thrilling Adventure Novels. Hearts of Three, by Jack London. The Crystal Skull, by Jack McLaren. The Master of Merripit, by Eden Phillpotts. Shadows by the Sea, by J. Jefferson Farjeon. [With illustrations.]
Record 78: (a) [Symphony] :
Record 79: (a) Thomas Hardy's novel universe :
Record 80: (a) Our present hope and our future home :
Record 81: (a) The paradox of prosperity :
Record 82: (a) The Lancashire local historian and his theme.
Record 83: (a) Business and government :
Record 84: (a) Gamma rays from isotopes produced by (n, gamma) - reactions /
Record 85: (a) Dante - the Divine comedy :`

	data := openTestMARC(t)
	defer data.Close()

	count := 0
	titles := []string{}
	for {
		r, err := ReadRecord(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		count++
		t := r.GetSubFields("245", 'a')
		if len(t) != 0 {
			for _, f := range t {
				titles = append(titles, fmt.Sprintf("Record %d: %s", count, f))
			}
		}
	}
	out := strings.Join(titles, "\n")
	if out != exp {
		t.Errorf("Returned subfields %s did not match expected subfields", out)
	}
	log.Printf("GetSubFields()")
}

// Basic example of how to modify a record
func TestAppendField(t *testing.T) {
	const exp = `001 50001
005 20010903131819.0
008 701012s1970    moua     b    001 0 eng  
010 [  ] [(a)    73117956 ]
035 [  ] [(a) ocm00094426 ]
035 [  ] [(9) 7003024381]
040 [  ] [(a) DLC], [(c) DLC], [(d) OKO]
020 [  ] [(a) 0801657024]
050 [00] [(a) RC78.7.C9], [(b) Z83]
060 [  ] [(a) QS 504 Z94d 1970]
082 [00] [(a) 616.07/583]
049 [  ] [(a) CUDA]
100 [1 ] [(a) Zugibe, Frederick T.], [(q) (Frederick Thomas),], [(d) 1928-]
245 [10] [(a) Diagnostic histochemistry], [(c) [by] Frederick T. Zugibe.]
260 [  ] [(a) Saint Louis,], [(b) Mosby,], [(c) 1970.]
300 [  ] [(a) xiv, 366 p.], [(b) illus.], [(c) 25 cm.]
504 [  ] [(a) Bibliography: p. 332-349.]
650 [ 0] [(a) Cytodiagnosis.]
650 [ 0] [(a) Histochemistry], [(x) Technique.]
650 [ 2] [(a) Histocytochemistry.]
650 [ 2] [(a) Histological Techniques.]
994 [  ] [(a) 92], [(b) CUD]
650 [ 0] [(a) Unit testing.]`

	data := openTestMARC(t)
	defer data.Close()

	r, _ := ReadRecord(data)
	// Build a subfield first
	s := &SubField{Code: 'a', Value: `Unit testing.`}
	// Wrap the subfield in a slice
	sf := make([]*SubField, 0)
	sf = append(sf, s)
	// Create a datafield containing the subfield slice
	df := &DataField{Tag: `650`, Ind1: ' ', Ind2: '0', SubFields: sf}
	// Append the datafield to the record
	r.Fields = append(r.Fields, df)
	out := r.String()
	if out != exp {
		t.Errorf("Output record string %s did not match expected string", out)
	}
	log.Printf("Append a field to a record")
}
