
package lucy.deps;
import java.lang.reflect.* ; 

public class ArrayInt   {
	public int start;
	public int end; // not include
	public int cap;
	static String outOfRagneMsg = "index out range";
	public int[] elements;
	public int size(){
		return this.end - this.start;
	}
	public int start(){
        return this.start;
	}
	public int end(){
         return this.end;
	}
	public int cap(){
         return this.end;
	}
	public ArrayInt(int[] values){
		this.start = 0;
		this.end = values.length;
		this.cap = values.length;
		this.elements = values;
		
	}
	private ArrayInt(){

	}
	public ArrayInt slice(int start,int end){
		if(end  < 0 ){
		      end = this.end - this.start;  // whole length
		}
		if(start < 0 || start > end || end + this.start > this.end){
			throw new ArrayIndexOutOfBoundsException(outOfRagneMsg);
		}
		ArrayInt result = new ArrayInt();
		result.elements = this.elements;
		result.start = this.start + start;
		result.end = this.start + end;
		result.cap = this.cap;
		return result;
	}
	public ArrayInt append(int e){
		if(this.end < this.cap){
		}else{
			this.expand(this.cap * 2);
		}
		this.elements[this.end++] = e;
		return this;
	}
	private void expand(int cap){
		if(cap <= 0){
		    cap = 10;
		}
		Class c = this.elements.getClass();
		int[] eles = (int[]) Array.newInstance(c.getComponentType() , cap );
		int length = this.size();
		for(int i = 0;i < length;i++){
			eles[i] = this.elements[i + this.start];
		}
		this.start = 0;
		this.end = length;
		this.cap = cap;
		this.elements = eles;
	}
	public ArrayInt append(ArrayInt es){
		if(this.end + es.size() < this.cap){
		}else {
			this.expand((this.cap + es.size()) * 2);
		}
		for(int i = 0;i < es.size();i++){
			this.elements[this.end + i] = es.elements[es.start + i ];
		}
		this.end += es.size();
		return this;
	}
	public String toString(){
	    String s = "[";
	    int size = this.end - this.start;
	    for(int i= 0;i < size;i ++){
            s += this.elements[this.start + i ];
            if(i != size -1){
                s += " ";
            }
	    }
	    s += "]";
	    return s;
	}
	public int[] getJavaArray(){
		if(this.start == 0 && this.end == this.cap){
			return this.elements;
		}
		int length = this.end - this.start;
		int[] elements = new int[length];
		for(int i = 0; i < length; i ++){
			elements[i] = this.elements[i + this.start];
		}
		this.start = 0;
		this.end = length;
		this.elements = elements;
		this.cap = length;
		return elements;
	}
}

